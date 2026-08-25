package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config representa a configuração da aplicação
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Services  ServicesConfig  `mapstructure:"services"`
	RabbitMQ  RabbitMQConfig  `mapstructure:"rabbitmq"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Log       LogConfig       `mapstructure:"log"`
	Redis     RedisConfig     `mapstructure:"redis"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Env       string          `mapstructure:"env"`
}

// RedisConfig configurações de conexão com Redis
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RateLimitConfig configurações gerais de Rate Limit
type RateLimitConfig struct {
	Enabled bool                  `mapstructure:"enabled"`
	Rules   RateLimitRulesConfig  `mapstructure:"rules"`
}

// RateLimitRulesConfig mapeamento das regras específicas
type RateLimitRulesConfig struct {
	Login                 RateLimitRule `mapstructure:"login"`
	ForgotPassword        RateLimitRule `mapstructure:"forgot_password"`
	DeviceChallengeSend   RateLimitRule `mapstructure:"device_challenge_send"`
	DeviceChallengeVerify RateLimitRule `mapstructure:"device_challenge_verify"`
	ResetPassword         RateLimitRule `mapstructure:"reset_password"`
	ChangePassword        RateLimitRule `mapstructure:"change_password"`
	Refresh               RateLimitRule `mapstructure:"refresh"`
	Validate              RateLimitRule `mapstructure:"validate"`
	Logout                RateLimitRule `mapstructure:"logout"`
	DeviceQuickRevoke     RateLimitRule `mapstructure:"device_quick_revoke"`
	DeviceBlacklist       RateLimitRule `mapstructure:"device_blacklist"`
	Default               RateLimitRule `mapstructure:"default"`
}

// RateLimitRule define o limite e a janela de tempo
type RateLimitRule struct {
	Limit  int           `mapstructure:"limit"`
	Window time.Duration `mapstructure:"window"`
}

// ServerConfig configurações do servidor HTTP
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// ServicesConfig configurações dos microserviços
type ServicesConfig struct {
	Auth          ServiceConfig `mapstructure:"auth"`
	User          ServiceConfig `mapstructure:"user"`
	Company       ServiceConfig `mapstructure:"company"`
	Terms         ServiceConfig `mapstructure:"terms"`
	Communication ServiceConfig `mapstructure:"communication"`
}

// ServiceConfig configurações de um microserviço
type ServiceConfig struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retries int           `mapstructure:"retries"`
}

// RabbitMQConfig configurações do RabbitMQ
type RabbitMQConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	VHost      string `mapstructure:"vhost"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
	Durable    bool   `mapstructure:"durable"`
	AutoDelete bool   `mapstructure:"auto_delete"`
}

// JWTConfig configurações JWT
type JWTConfig struct {
	Issuer     string `mapstructure:"issuer"`
	Audience   string `mapstructure:"audience"`
	JWKSURL    string `mapstructure:"jwks_url"`
	SigningKey string `mapstructure:"signing_key"`
}

// MetricsConfig configurações de métricas
type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	ScrapePath string `mapstructure:"scrape_path"`
	Port       string `mapstructure:"port"`
}

// LogConfig configurações de log
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load carrega a configuração
func Load() (*Config, error) {
	// Configurações padrão
	setDefaults()

	// Configurações de ambiente
	viper.SetEnvPrefix("BFF_AUTH")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Determina o ambiente
	env := viper.GetString("env")
	if env == "" {
		env = "local"
	}

	// Carrega arquivo específico do ambiente
	viper.SetConfigName("application-" + env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/bff-auth")

	// Lê o arquivo de configuração
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults define valores padrão
func setDefaults() {
	// Server
	viper.SetDefault("server.port", "8381")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")

	// Services
	viper.SetDefault("services.auth.base_url", "http://localhost:8081")
	viper.SetDefault("services.auth.timeout", "5s")
	viper.SetDefault("services.auth.retries", 3)

	viper.SetDefault("services.user.base_url", "http://localhost:8082")
	viper.SetDefault("services.user.timeout", "5s")
	viper.SetDefault("services.user.retries", 3)

	viper.SetDefault("services.company.base_url", "http://localhost:8083")
	viper.SetDefault("services.company.timeout", "5s")
	viper.SetDefault("services.company.retries", 3)

	viper.SetDefault("services.terms.base_url", "http://localhost:8084")
	viper.SetDefault("services.terms.timeout", "5s")
	viper.SetDefault("services.terms.retries", 3)

	viper.SetDefault("services.communication.base_url", "http://localhost:8085")
	viper.SetDefault("services.communication.timeout", "5s")
	viper.SetDefault("services.communication.retries", 3)

	// RabbitMQ
	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", 5672)
	viper.SetDefault("rabbitmq.user", "guest")
	viper.SetDefault("rabbitmq.password", "guest")
	viper.SetDefault("rabbitmq.vhost", "/")
	viper.SetDefault("rabbitmq.exchange", "ms-communication-exchange-local")
	viper.SetDefault("rabbitmq.routing_key", "communication.message.send")
	viper.SetDefault("rabbitmq.durable", true)
	viper.SetDefault("rabbitmq.auto_delete", false)

	// JWT
	viper.SetDefault("jwt.issuer", "keepguard")
	viper.SetDefault("jwt.audience", "keepguard-api")

	// Metrics
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.scrape_path", "/metrics")
	viper.SetDefault("metrics.port", "9092")

	// Log
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	// Environment
	viper.SetDefault("env", "dev")
}
