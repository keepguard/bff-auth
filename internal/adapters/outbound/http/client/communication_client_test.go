package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestNewCommunicationClient testa a criação do CommunicationClient
func TestNewCommunicationClient(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Services: config.ServicesConfig{
			Communication: config.ServiceConfig{
				BaseURL: "http://localhost:8085",
				Timeout: 30 * time.Second,
				Retries: 3,
			},
		},
	}
	logger, _ := zap.NewDevelopment()

	// Act
	client := NewCommunicationClient(cfg, logger)

	// Assert
	assert.NotNil(t, client)
	commClient, ok := client.(*communicationClient)
	assert.True(t, ok)
	assert.NotNil(t, commClient.httpClient)
	assert.NotNil(t, commClient.config)
	assert.NotNil(t, commClient.logger)
}

// TestNewCommunicationClientWithLogger testa a criação do CommunicationClient com logger.Logger
func TestNewCommunicationClientWithLogger(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Services: config.ServicesConfig{
			Communication: config.ServiceConfig{
				BaseURL: "http://localhost:8085",
				Timeout: 30 * time.Second,
				Retries: 3,
			},
		},
	}

	// Act
	client := NewCommunicationClientWithLogger(cfg, nil)

	// Assert
	assert.NotNil(t, client)
}

// TestCommunicationClient_SendMessage_Success testa envio de mensagem com sucesso
func TestCommunicationClient_SendMessage_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar método e path
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/messages/send", r.URL.Path)
		assert.Equal(t, "test-correlation", r.Header.Get("X-Correlation-ID"))
		assert.Equal(t, "test-app", r.Header.Get("X-Application"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verificar body
		var req outboundDto.SendMessageRequestDTO
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "EMAIL", req.MessageType)
		assert.Equal(t, "EMAIL", req.CommunicationType)
		assert.Equal(t, "RECUPERACAO_SENHA", req.TemplateType)
		assert.Equal(t, "test@example.com", req.Recipient)
		assert.Equal(t, "TestUser", req.Variables["userName"])
		assert.Equal(t, "555555", req.Variables["token"])

		// Resposta de sucesso
		response := outboundDto.SendMessageResponseDTO{
			Success: true,
			Message: "Mensagem enviada com sucesso",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Communication: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	client := NewCommunicationClient(cfg, logger)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Recipient:         "test@example.com",
		Variables: map[string]interface{}{
			"userName": "TestUser",
			"token":    "555555",
		},
	}
	xApplication := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.SendMessage(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "Mensagem enviada com sucesso", result.Message)
}

// TestCommunicationClient_SendMessage_HTTPError testa envio de mensagem com erro HTTP estruturado
func TestCommunicationClient_SendMessage_HTTPError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resposta de erro HTTP 400
		msError := appdto.MSAuthErrorResponse{
			Type:   "about:blank",
			Title:  "Bad Request",
			Status: 400,
			Detail: "Template não encontrado",
			Properties: map[string]interface{}{
				"errorCode": "TEMPLATE_NOT_FOUND",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msError)
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Communication: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	client := NewCommunicationClient(cfg, logger)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "INVALID_TEMPLATE",
		Recipient:         "test@example.com",
		Variables:         map[string]interface{}{},
	}
	xApplication := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.SendMessage(ctx, req, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, outboundDto.SendMessageResponseDTO{}, result)

	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Equal(t, "Template não encontrado", httpErr.Message)
	assert.Equal(t, "TEMPLATE_NOT_FOUND", httpErr.ErrorCode)
}

// TestCommunicationClient_SendMessage_GenericError testa envio de mensagem com erro genérico
func TestCommunicationClient_SendMessage_GenericError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resposta de erro genérico
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Communication: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	client := NewCommunicationClient(cfg, logger)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Recipient:         "test@example.com",
		Variables:         map[string]interface{}{},
	}
	xApplication := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.SendMessage(ctx, req, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, outboundDto.SendMessageResponseDTO{}, result)

	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Equal(t, "erro de send message: status 500", httpErr.Message)
	assert.Equal(t, "HTTP_ERROR", httpErr.ErrorCode)
}

// TestCommunicationClient_SendMessage_RequestError testa envio de mensagem com erro de requisição
func TestCommunicationClient_SendMessage_RequestError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simular timeout - não responder
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Communication: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}
	logger, _ := zap.NewDevelopment()

	// Criar client manualmente para poder modificar timeout
	httpClient := resty.New()
	httpClient.SetTimeout(100 * time.Millisecond)
	httpClient.SetRetryCount(0) // Desabilita retry para teste mais rápido

	commClient := &communicationClient{
		httpClient: httpClient,
		config:     cfg,
		logger:     logger,
	}

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Recipient:         "test@example.com",
		Variables:         map[string]interface{}{},
	}
	xApplication := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := commClient.SendMessage(ctx, req, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, outboundDto.SendMessageResponseDTO{}, result)
	assert.Contains(t, err.Error(), "erro ao comunicar com ms-communication")
}

// TestCommunicationClient_HandleHTTPError_Success testa handleHTTPError com status de sucesso
func TestCommunicationClient_HandleHTTPError_Success(t *testing.T) {
	// Arrange
	client := &communicationClient{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	restyClient := resty.New()
	resp, err := restyClient.R().Get(server.URL)
	assert.NoError(t, err)

	// Act
	err = client.handleHTTPError(resp, "test")

	// Assert
	assert.NoError(t, err)
}

// TestCommunicationClient_HandleHTTPError_MSCommunicationError testa handleHTTPError com erro estruturado
func TestCommunicationClient_HandleHTTPError_MSCommunicationError(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	client := &communicationClient{
		logger: logger,
	}

	msError := appdto.MSAuthErrorResponse{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: 400,
		Detail: "Dados inválidos",
		Properties: map[string]interface{}{
			"errorCode": "INVALID_DATA",
		},
	}
	msErrorJSON, _ := json.Marshal(msError)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(msErrorJSON)
	}))
	defer server.Close()

	restyClient := resty.New()
	resp, err := restyClient.R().Get(server.URL)
	assert.NoError(t, err)

	// Act
	err = client.handleHTTPError(resp, "test")

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Equal(t, "Dados inválidos", httpErr.Message)
	assert.Equal(t, "INVALID_DATA", httpErr.ErrorCode)
}

// TestCommunicationClient_HandleHTTPError_MSCommunicationErrorWithoutErrorCode testa com Properties sem errorCode
func TestCommunicationClient_HandleHTTPError_MSCommunicationErrorWithoutErrorCode(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	client := &communicationClient{
		logger: logger,
	}

	msError := appdto.MSAuthErrorResponse{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: 400,
		Detail: "Erro desconhecido",
		Properties: map[string]interface{}{
			"timestamp": "2023-01-01T00:00:00Z",
		},
	}
	msErrorJSON, _ := json.Marshal(msError)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(msErrorJSON)
	}))
	defer server.Close()

	restyClient := resty.New()
	resp, err := restyClient.R().Get(server.URL)
	assert.NoError(t, err)

	// Act
	err = client.handleHTTPError(resp, "test")

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Equal(t, "Erro desconhecido", httpErr.Message)
	assert.Equal(t, "UNKNOWN_ERROR", httpErr.ErrorCode)
}

// TestCommunicationClient_HandleHTTPError_GenericError testa handleHTTPError com erro genérico
func TestCommunicationClient_HandleHTTPError_GenericError(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	client := &communicationClient{
		logger: logger,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	restyClient := resty.New()
	resp, err := restyClient.R().Get(server.URL)
	assert.NoError(t, err)

	// Act
	err = client.handleHTTPError(resp, "test")

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Equal(t, "erro de test: status 500", httpErr.Message)
	assert.Equal(t, "HTTP_ERROR", httpErr.ErrorCode)
	assert.Equal(t, "Internal Server Error", httpErr.Details["body"])
}

// TestCommunicationClient_HandleHTTPError_InvalidJSON testa handleHTTPError com JSON inválido
func TestCommunicationClient_HandleHTTPError_InvalidJSON(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	client := &communicationClient{
		logger: logger,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	restyClient := resty.New()
	resp, err := restyClient.R().Get(server.URL)
	assert.NoError(t, err)

	// Act
	err = client.handleHTTPError(resp, "test")

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Equal(t, "erro de test: status 400", httpErr.Message)
	assert.Equal(t, "HTTP_ERROR", httpErr.ErrorCode)
}

// TestCommunicationClient_HandleHTTPError_EmptyTitle testa handleHTTPError com título vazio
func TestCommunicationClient_HandleHTTPError_EmptyTitle(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	client := &communicationClient{
		logger: logger,
	}

	msError := appdto.MSAuthErrorResponse{
		Type:   "about:blank",
		Title:  "", // Título vazio
		Status: 400,
		Detail: "Bad Request",
	}
	msErrorJSON, _ := json.Marshal(msError)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(msErrorJSON)
	}))
	defer server.Close()

	restyClient := resty.New()
	resp, err := restyClient.R().Get(server.URL)
	assert.NoError(t, err)

	// Act
	err = client.handleHTTPError(resp, "test")

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Equal(t, "erro de test: status 400", httpErr.Message)
	assert.Equal(t, "HTTP_ERROR", httpErr.ErrorCode)
}
