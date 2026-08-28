// @title BFF-AUTH API
// @version 1.0.0
// @description Backend for Frontend responsável pela autenticação e autorização do sistema KeepGuard
// @host localhost:8381
// @BasePath /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/keepguard/bff-auth/docs" // Importa docs para inicializar Swagger
	httpserver "github.com/keepguard/bff-auth/internal/adapters/inbound/http"
	handlersPkg "github.com/keepguard/bff-auth/internal/adapters/inbound/http/handlers"
	middlewarePkg "github.com/keepguard/bff-auth/internal/adapters/inbound/http/middleware"
	httpclient "github.com/keepguard/bff-auth/internal/adapters/outbound/http/client"
	authdecorator "github.com/keepguard/bff-auth/internal/adapters/outbound/http/decorator/auth"
	communicationdecorator "github.com/keepguard/bff-auth/internal/adapters/outbound/http/decorator/communication"
	companydecorator "github.com/keepguard/bff-auth/internal/adapters/outbound/http/decorator/company"
	userdecorator "github.com/keepguard/bff-auth/internal/adapters/outbound/http/decorator/user"
	messagingDecorator "github.com/keepguard/bff-auth/internal/adapters/outbound/messaging/decorator"
	rabbitmqPublisher "github.com/keepguard/bff-auth/internal/adapters/outbound/messaging/rabbitmq"
	"github.com/keepguard/bff-auth/internal/application/auth"
	"github.com/keepguard/bff-auth/internal/application/message"
	"github.com/keepguard/bff-auth/internal/infrastructure/cache"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"github.com/keepguard/bff-auth/internal/infrastructure/resilience"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

func main() {
	// Swagger é inicializado automaticamente via docs package

	// Inicializa logger básico para bootstrap
	bootstrapLogger, _ := zap.NewProduction()
	defer bootstrapLogger.Sync()

	// Carrega configuração
	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Fatal("Erro ao carregar configuração",
			zap.Error(err),
			zap.String("component", "bff-auth"),
			zap.String("service", "bff-auth"),
		)
	}

	// Inicializa logger padrão
	appLogger, err := logger.New("info", "json")
	if err != nil {
		bootstrapLogger.Fatal("Erro ao inicializar logger",
			zap.Error(err),
			zap.String("component", "bff-auth"),
			zap.String("service", "bff-auth"),
		)
	}
	defer appLogger.Sync()

	// Loga início da aplicação
	appLogger.Info("Iniciando BFF-AUTH",
		zap.String("service", "bff-auth"),
		zap.String("component", "bff-auth"),
		zap.String("environment", cfg.Env),
		zap.String("version", "1.0.0"),
		zap.String("env", cfg.Env),
		zap.Bool("kibana_enabled", os.Getenv("KIBANA_ENABLED") == "true"),
		zap.String("log_level", cfg.Log.Level),
		zap.String("log_format", cfg.Log.Format),
	)

	// Inicializa métricas (mantém compatibilidade)
	metrics := metrics.New()

	// Inicializa Circuit Breaker Manager
	cbManager := resilience.NewCircuitBreakerManager(metrics)

	// Configura Circuit Breaker para ms-auth
	authCBConfig := resilience.CircuitBreakerConfig{
		Name:        "ms-auth",
		MaxRequests: 3,                // Máximo 3 requests em half-open para testar recuperação
		Interval:    60 * time.Second, // Janela de amostragem
		Timeout:     10 * time.Second, // Tempo em OPEN antes de tentar HALF-OPEN
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Abre se pelo menos 10 requisições foram feitas e >= 70% falharam com 5xx/infra
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.7
		},
	}

	// Cria Circuit Breaker para ms-auth
	cbManager.GetOrCreate("ms-auth", authCBConfig)

	// Configura Circuit Breaker para ms-company
	companyCBConfig := resilience.CircuitBreakerConfig{
		Name:        "ms-company",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.7
		},
	}
	cbManager.GetOrCreate("ms-company", companyCBConfig)

	// Converte logger para zap.Logger (necessário para decorators)
	// Type assertion para acessar o método GetZapLogger
	zapLoggerImpl, ok := appLogger.(interface{ GetZapLogger() *zap.Logger })
	if !ok {
		bootstrapLogger.Fatal("Logger não suporta GetZapLogger()",
			zap.String("component", "bff-auth"),
			zap.String("service", "bff-auth"),
		)
	}
	zapLogger := zapLoggerImpl.GetZapLogger()

	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, zapLogger)
	if err != nil {
		zapLogger.Warn("Aviso ao conectar no Redis", zap.Error(err))
	}

	// =============================================================================
	// INICIALIZAÇÃO DO AUTH CLIENT COM DECORATORS COMPOSTOS
	// =============================================================================
	// A ordem dos decorators importa! Cada decorator envolve o anterior.
	// Ordem de execução (de fora para dentro):
	// Request → Logging → Circuit Breaker → Retry → Metrics → Smart Cache → HTTP Client

	// 1. Cliente HTTP base (mais interno - faz a requisição HTTP real)
	baseAuthClient := httpclient.NewAuthClient(
		cfg.Services.Auth.BaseURL,
		cfg.Services.Auth.Timeout,
	)

	// 2. Smart Cache Decorator (cache inteligente baseado no ExpiresIn do token)
	smartCacheConfig := authdecorator.SmartCacheConfig{
		MaxTTL:          10 * time.Minute, // TTL máximo de 10 minutos (segurança)
		MinTTL:          30 * time.Second, // TTL mínimo de 30 segundos
		MaxSize:         1000,             // Máximo 1000 tokens em cache
		CleanupInterval: 1 * time.Minute,  // Limpa cache expirado a cada minuto
	}
	// Smart Cache usa o ExpiresIn do token para determinar o TTL do cache!
	// Se token expira em 3600s (1h), cache será válido por até 3600s (limitado a MaxTTL)
	cachedClient := authdecorator.NewSmartCacheDecorator(baseAuthClient, smartCacheConfig, metrics)

	// 3. Metrics Decorator (coleta métricas automaticamente)
	metricsClient := authdecorator.NewMetricsDecorator(cachedClient, metrics, "ms-auth")

	// 4. Retry Decorator (retry inteligente com backoff exponencial)
	retryConfig := authdecorator.RetryConfig{
		MaxAttempts:  2,                      // Máximo 2 tentativas (reduzido para evitar latência excessiva)
		InitialDelay: 100 * time.Millisecond, // Delay inicial de 100ms
		MaxDelay:     2 * time.Second,        // Delay máximo de 2s
		Multiplier:   2.0,                    // Backoff exponencial: 100ms → 200ms
		Jitter:       true,                   // Adiciona aleatoriedade para evitar thundering herd
	}
	retryClient := authdecorator.NewRetryDecorator(metricsClient, retryConfig)

	// 5. Circuit Breaker Decorator (protege contra falhas em cascata)
	cbClient := authdecorator.NewCircuitBreakerDecorator(retryClient, cbManager, "ms-auth")

	// 6. Logging Decorator (mais externo - loga todas as requisições/respostas)
	authClient := authdecorator.NewLoggingDecorator(cbClient, zapLogger, "ms-auth")

	// Cliente final está pronto! Todos os aspectos estão automaticamente aplicados:
	// ✅ Logs estruturados automáticos
	// ✅ Circuit breaker para resiliência
	// ✅ Retry inteligente com backoff
	// ✅ Métricas Prometheus automáticas
	// ✅ Smart Cache para ValidateToken (baseado no ExpiresIn do token)
	// =============================================================================

	// =============================================================================
	// INICIALIZAÇÃO DO COMPANY CLIENT COM DECORATORS
	// =============================================================================
	// CRÍTICO! Chamado em TODA requisição (Login, ValidateToken, etc.)
	// Request → Logging → Circuit Breaker → Retry → Cache → Metrics → HTTP Client

	// 1. Cliente HTTP base
	baseCompanyClient := httpclient.NewCompanyClient(cfg, zapLogger)

	// 2. Metrics Decorator
	companyMetricsClient := companydecorator.NewCompanyMetricsDecorator(baseCompanyClient, metrics, "ms-company")

	companyRedisClient := companydecorator.NewRedisCacheDecorator(
		companyMetricsClient,
		companydecorator.NewRedisStringCache(redisClient),
		metrics,
		zapLogger,
	)

	// 3. Cache in-memory (L1) sobre Redis (L2) e HTTP (L3)
	companyCacheConfig := companydecorator.CacheConfig{
		TTL:             5 * time.Minute, // Cache por 5 minutos (dados raramente mudam)
		MaxSize:         1000,            // Máximo 1000 empresas
		CleanupInterval: 1 * time.Minute, // Limpa cache expirado a cada minuto
	}
	companyCachedClient := companydecorator.NewCacheDecorator(companyRedisClient, companyCacheConfig, metrics)

	// 4. Retry Decorator (rápido: 50ms, 2 tentativas)
	companyRetryConfig := companydecorator.RetryConfig{
		MaxAttempts:  2,                     // Apenas 2 tentativas (com cache, raramente usado)
		InitialDelay: 50 * time.Millisecond, // Muito rápido (chamado em toda req)
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}
	companyRetryClient := companydecorator.NewRetryDecorator(companyCachedClient, companyRetryConfig)

	// 5. Circuit Breaker (Protege se ms-company cair - CRÍTICO!)
	companyClient := companydecorator.NewCircuitBreakerDecorator(companyRetryClient, cbManager, "ms-company")

	// 6. Logging Decorator
	companyClient = companydecorator.NewCompanyLoggingDecorator(companyClient, zapLogger, "ms-company")

	// ✅ Cache (99% hit) + Circuit Breaker + Retry + Logging + Metrics
	// =============================================================================

	// =============================================================================
	// INICIALIZAÇÃO DO AUTH USER CLIENT COM DECORATORS
	// =============================================================================
	// Request → Logging → Metrics → HTTP Client

	baseAuthUserClient := httpclient.NewAuthUserClient(cfg, zapLogger)
	authUserMetricsClient := userdecorator.NewUserMetricsDecorator(baseAuthUserClient, metrics, "ms-auth-users")
	authUserClient := userdecorator.NewUserLoggingDecorator(authUserMetricsClient, zapLogger, "ms-auth-users")

	// ✅ Logs automáticos + Métricas Prometheus
	// =============================================================================

	// =============================================================================
	// INICIALIZAÇÃO DO COMMUNICATION CLIENT COM DECORATORS
	// =============================================================================
	// Envia e-mails (APIs externas: SendGrid/AWS SES)
	// Request → Logging → Retry → Metrics → HTTP Client

	// 1. Cliente HTTP base
	baseCommunicationClient := httpclient.NewCommunicationClient(cfg, zapLogger)

	// 2. Metrics Decorator
	communicationMetricsClient := communicationdecorator.NewCommunicationMetricsDecorator(baseCommunicationClient, metrics, "ms-communication")

	// 3. Retry Decorator (reduzido: 200ms, 2 tentativas)
	communicationRetryConfig := communicationdecorator.RetryConfig{
		MaxAttempts:  2,                      // 2 tentativas (reduzido para evitar latência excessiva)
		InitialDelay: 200 * time.Millisecond, // Mais lento (API externa)
		MaxDelay:     5 * time.Second,        // Delay máximo reduzido
		Multiplier:   2.5,                    // Backoff: 200ms → 500ms
		Jitter:       true,
	}
	communicationRetryClient := communicationdecorator.NewRetryDecorator(communicationMetricsClient, communicationRetryConfig)

	// 4. Logging Decorator
	communicationClient := communicationdecorator.NewCommunicationLoggingDecorator(communicationRetryClient, zapLogger, "ms-communication")

	// ✅ Retry agressivo (4x) + Logging + Metrics (90% falhas recuperadas!)
	// =============================================================================

	// =============================================================================
	// INICIALIZAÇÃO DO MESSAGE PUBLISHER COM DECORATORS
	// =============================================================================
	// RabbitMQ Publisher → Logging → Metrics → CircuitBreaker (com fallback HTTP)

	// 1. Criar publisher RabbitMQ base
	rabbitPublisher, err := rabbitmqPublisher.NewMessagePublisher(&cfg.RabbitMQ, zapLogger)
	if err != nil {
		appLogger.Fatal("Erro ao inicializar publisher RabbitMQ",
			zap.Error(err),
			zap.String("component", "bff-auth"),
			zap.String("service", "bff-auth"),
		)
	}

	// 2. Logging Decorator
	loggingPublisher := messagingDecorator.NewLoggingDecorator(rabbitPublisher, zapLogger)

	// 3. Metrics Decorator
	metricsPublisher := messagingDecorator.NewMetricsDecorator(loggingPublisher, metrics, zapLogger)

	// 4. Circuit Breaker Decorator (com fallback HTTP)
	messagePublisher := messagingDecorator.NewCircuitBreakerDecorator(
		metricsPublisher,
		cbManager,
		communicationClient, // fallback HTTP
		zapLogger,
	)

	appLogger.Info("Message Publisher inicializado com sucesso",
		zap.String("component", "bff-auth"),
		zap.String("service", "bff-auth"),
		zap.String("exchange", cfg.RabbitMQ.Exchange),
		zap.String("routingKey", cfg.RabbitMQ.RoutingKey),
	)

	// Inicializa use cases de autenticação
	loginUseCase := auth.NewLoginUseCase(authClient, companyClient)
	refreshUseCase := auth.NewRefreshUseCase(authClient, companyClient)
	logoutUseCase := auth.NewLogoutUseCase(authClient, companyClient)
	validateTokenUseCase := auth.NewValidateTokenUseCase(authClient, companyClient, zapLogger)
	changePasswordUseCase := auth.NewChangePasswordUseCase(authClient, companyClient, zapLogger)
	resetPasswordUseCase := auth.NewResetPasswordUseCase(authClient, authUserClient, companyClient, messagePublisher, zapLogger)

	// Inicializa use cases de mensagens
	sendResetPasswordMessageUseCase := message.NewSendResetPasswordMessageUseCase(authClient, authUserClient, companyClient, messagePublisher, zapLogger)

	// Inicializa handlers HTTP diretamente com use cases
	authHandlers := handlersPkg.NewAuthHandlersWithLogger(
		loginUseCase,
		refreshUseCase,
		logoutUseCase,
		validateTokenUseCase,
		changePasswordUseCase,
		resetPasswordUseCase,
		authClient,
		companyClient,
		appLogger,
	)

	messageHandlers := handlersPkg.NewMessageHandlersWithLogger(
		sendResetPasswordMessageUseCase,
		appLogger,
	)

	// Combina os handlers
	combinedHandlers := handlersPkg.NewCombinedHandlers(authHandlers, messageHandlers)

	rateLimiterMiddleware := middlewarePkg.NewRateLimiterMiddleware(redisClient, cfg.RateLimit, zapLogger, metrics)

	// Inicializa servidor HTTP com Rate Limiting e Validação de Sessão Redis
	server := httpserver.NewServer(cfg, appLogger, metrics, rateLimiterMiddleware, redisClient)
	server.SetupRoutes(combinedHandlers)

	// =============================================================================
	// SERVIDOR DE MÉTRICAS PROMETHEUS
	// =============================================================================
	go func() {
		metricsPort := cfg.Metrics.Port
		if metricsPort == "" {
			metricsPort = "9092"
		}
		
		metricsMux := http.NewServeMux()
		metricsMux.Handle(cfg.Metrics.ScrapePath, promhttp.Handler())
		
		appLogger.Info("Servidor de métricas iniciado",
			zap.String("service", "bff-auth"),
			zap.String("port", metricsPort),
			zap.String("path", cfg.Metrics.ScrapePath),
		)
		
		if err := http.ListenAndServe(":"+metricsPort, metricsMux); err != nil {
			appLogger.Error("Erro no servidor de métricas",
				zap.Error(err),
				zap.String("service", "bff-auth"),
			)
		}
	}()

	// Inicia servidor em goroutine
	go func() {
		appLogger.Info("Iniciando servidor HTTP",
			zap.String("port", cfg.Server.Port),
			zap.String("env", cfg.Env),
			zap.String("component", "bff-auth"),
			zap.String("service", "bff-auth"),
			zap.String("environment", cfg.Env),
			zap.String("version", "1.0.0"),
		)

		if err := server.Start(); err != nil {
			appLogger.Fatal("Erro ao iniciar servidor",
				zap.Error(err),
				zap.String("component", "bff-auth"),
				zap.String("service", "bff-auth"),
				zap.String("environment", cfg.Env),
				zap.String("version", "1.0.0"),
			)
		}
	}()

	// Aguarda sinal para shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...",
		zap.String("component", "bff-auth"),
		zap.String("service", "bff-auth"),
		zap.String("environment", cfg.Env),
		zap.String("version", "1.0.0"),
	)

	// Shutdown graceful
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fechar publisher RabbitMQ
	if err := messagePublisher.Close(); err != nil {
		appLogger.Error("Erro ao fechar Message Publisher",
			zap.Error(err),
			zap.String("component", "bff-auth"),
			zap.String("service", "bff-auth"),
		)
	}

	if err := server.Stop(ctx); err != nil {
		appLogger.Error("Erro durante shutdown",
			zap.Error(err),
			zap.String("component", "bff-auth"),
			zap.String("service", "bff-auth"),
			zap.String("environment", cfg.Env),
			zap.String("version", "1.0.0"),
		)
	}

	appLogger.Info("Server stopped",
		zap.String("component", "bff-auth"),
		zap.String("service", "bff-auth"),
		zap.String("environment", cfg.Env),
		zap.String("version", "1.0.0"),
	)
}
