package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authctx "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
)

// TestNewAuthClient testa a criação do AuthClient
func TestNewAuthClient(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8081"
	timeout := 30 * time.Second

	// Act
	client := NewAuthClient(baseURL, timeout)

	// Assert
	assert.NotNil(t, client)
	authClient, ok := client.(*AuthClient)
	assert.True(t, ok)
	assert.Equal(t, baseURL, authClient.baseURL)
	assert.NotNil(t, authClient.client)
}

// TestAuthClient_Login_Success testa login com sucesso
func TestAuthClient_Login_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/auth/login", r.URL.Path)
		assert.Equal(t, "test-correlation", r.Header.Get("X-Correlation-ID"))
		assert.Empty(t, r.Header.Get("X-Tenant-Id"))

		// Verificar body
		var req inboundDto.AuthRequestDTO
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "testpass", req.Password)

		// Resposta de sucesso
		response := inboundDto.AuthResponseDTO{
			Token:     "access_token",
			ExpiresIn: 3600,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app"
	correlationID := "test-correlation"

	expectedResponse := inboundDto.AuthResponseDTO{
		Token:     "access_token",
		ExpiresIn: 3600,
	}

	// Act
	result, err := client.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
}

// TestAuthClient_Login_HTTPError testa login com erro HTTP
func TestAuthClient_Login_HTTPError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resposta de erro HTTP 401
		msError := appdto.MSAuthErrorResponse{
			Type:   "about:blank",
			Title:  "Unauthorized",
			Status: 401,
			Detail: "Credenciais inválidas",
			Properties: map[string]interface{}{
				"errorCode": "INVALID_CREDENTIALS",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(msError)
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)

	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 401, httpErr.StatusCode)
	assert.Equal(t, "Credenciais inválidas", httpErr.Message)
	assert.Equal(t, "INVALID_CREDENTIALS", httpErr.ErrorCode)
}

// TestAuthClient_Login_GenericError testa login com erro genérico
func TestAuthClient_Login_GenericError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resposta de erro genérico
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)

	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Equal(t, "erro de login: status 500", httpErr.Message)
	assert.Equal(t, "HTTP_ERROR", httpErr.ErrorCode)
}

// TestAuthClient_RefreshToken_Success testa refresh token com sucesso
func TestAuthClient_RefreshToken_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/auth/refresh", r.URL.Path)
		assert.Equal(t, "test-correlation", r.Header.Get("X-Correlation-ID"))
		assert.Equal(t, "test-company", r.Header.Get("X-Company-Id"))

		// Verificar body
		var req inboundDto.RefreshTokenRequestDTO
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "refresh_token", req.Token)

		// Resposta de sucesso
		response := inboundDto.RefreshTokenResponseDTO{
			Token:     "new_access_token",
			ExpiresIn: 3600,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	req := inboundDto.RefreshTokenRequestDTO{
		Token: "refresh_token",
	}
	tenantId := "test-app"
	correlationID := "test-correlation"

	expectedResponse := inboundDto.RefreshTokenResponseDTO{
		Token:     "new_access_token",
		ExpiresIn: 3600,
	}

	// Act
	result, err := client.RefreshToken(ctx, req, tenantId, correlationID, "keepguard-default-client")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
}

// TestAuthClient_RefreshToken_HTTPError testa refresh token com erro HTTP
func TestAuthClient_RefreshToken_HTTPError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resposta de erro HTTP 400
		msError := appdto.MSAuthErrorResponse{
			Type:   "about:blank",
			Title:  "Bad Request",
			Status: 400,
			Detail: "Refresh token inválido",
			Properties: map[string]interface{}{
				"errorCode": "INVALID_REFRESH_TOKEN",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msError)
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	req := inboundDto.RefreshTokenRequestDTO{
		Token: "invalid_refresh_token",
	}
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.RefreshToken(ctx, req, tenantId, correlationID, "keepguard-default-client")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.RefreshTokenResponseDTO{}, result)

	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Equal(t, "Refresh token inválido", httpErr.Message)
	assert.Equal(t, "INVALID_REFRESH_TOKEN", httpErr.ErrorCode)
}

// TestAuthClient_Logout_Success testa logout com sucesso
func TestAuthClient_Logout_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/auth/logout", r.URL.Path)
		assert.Equal(t, "Bearer access_token", r.Header.Get("Authorization"))
		assert.Equal(t, "test-correlation", r.Header.Get("X-Correlation-ID"))
		assert.Equal(t, "test-company", r.Header.Get("X-Company-Id"))

		// Resposta de sucesso
		response := map[string]string{"message": "Logout realizado com sucesso"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	token := "access_token"
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	err := client.Logout(ctx, token, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
}

// TestAuthClient_Logout_HTTPError testa logout com erro HTTP
func TestAuthClient_Logout_HTTPError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resposta de erro HTTP 401
		msError := appdto.MSAuthErrorResponse{
			Type:   "about:blank",
			Title:  "Unauthorized",
			Status: 401,
			Detail: "Token inválido",
			Properties: map[string]interface{}{
				"errorCode": "INVALID_TOKEN",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(msError)
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	token := "invalid_token"
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	err := client.Logout(ctx, token, tenantId, correlationID)

	// Assert
	assert.Error(t, err)

	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 401, httpErr.StatusCode)
	assert.Equal(t, "Token inválido", httpErr.Message)
	assert.Equal(t, "INVALID_TOKEN", httpErr.ErrorCode)
}

// TestAuthClient_Logout_GenericError testa logout com erro genérico
func TestAuthClient_Logout_GenericError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resposta de erro genérico
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewAuthClient(server.URL, 30*time.Second)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	token := "access_token"
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	err := client.Logout(ctx, token, tenantId, correlationID)

	// Assert
	assert.Error(t, err)

	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Equal(t, "erro de logout: status 500", httpErr.Message)
	assert.Equal(t, "HTTP_ERROR", httpErr.ErrorCode)
}

// TestHandleHTTPError_Success testa handleHTTPError com status de sucesso
func TestHandleHTTPError_Success(t *testing.T) {
	// Arrange
	client := &AuthClient{}

	// Criar um mock response com status 200 usando httptest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Fazer uma requisição real para obter um resty.Response válido
	restyClient := resty.New()
	resp, err := restyClient.R().Get(server.URL)
	assert.NoError(t, err)

	// Act
	err = client.handleHTTPError(resp, "test")

	// Assert
	assert.NoError(t, err)
}

// TestHandleHTTPError_MSAuthError testa handleHTTPError com erro estruturado do ms-auth
func TestHandleHTTPError_MSAuthError(t *testing.T) {
	// Arrange
	client := &AuthClient{}
	msError := appdto.MSAuthErrorResponse{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: 400,
		Detail: "Requisição inválida",
		Properties: map[string]interface{}{
			"errorCode": "INVALID_REQUEST",
			"timestamp": "2023-01-01T00:00:00Z",
		},
	}
	msErrorJSON, _ := json.Marshal(msError)

	// Criar um servidor que retorna erro 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(msErrorJSON)
	}))
	defer server.Close()

	// Fazer uma requisição real para obter um resty.Response válido
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
	assert.Equal(t, "Requisição inválida", httpErr.Message)
	assert.Equal(t, "INVALID_REQUEST", httpErr.ErrorCode)
	assert.Equal(t, msError.Properties, httpErr.Details)
}

// TestHandleHTTPError_GenericError testa handleHTTPError com erro genérico
func TestHandleHTTPError_GenericError(t *testing.T) {
	// Arrange
	client := &AuthClient{}

	// Criar um servidor que retorna erro 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Fazer uma requisição real para obter um resty.Response válido
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

// TestHandleHTTPError_InvalidJSON testa handleHTTPError com JSON inválido
func TestHandleHTTPError_InvalidJSON(t *testing.T) {
	// Arrange
	client := &AuthClient{}

	// Criar um servidor que retorna JSON inválido
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// Fazer uma requisição real para obter um resty.Response válido
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

// TestHandleHTTPError_EmptyTitle testa handleHTTPError com título vazio
func TestHandleHTTPError_EmptyTitle(t *testing.T) {
	// Arrange
	client := &AuthClient{}
	msError := appdto.MSAuthErrorResponse{
		Type:   "about:blank",
		Title:  "", // Título vazio
		Status: 400,
		Detail: "Bad Request",
	}
	msErrorJSON, _ := json.Marshal(msError)

	// Criar um servidor que retorna erro com título vazio
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(msErrorJSON)
	}))
	defer server.Close()

	// Fazer uma requisição real para obter um resty.Response válido
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
	assert.Equal(t, "Bad Request", httpErr.Message)
	assert.Equal(t, "UNKNOWN_ERROR", httpErr.ErrorCode)
}

// TestGetErrorCodeFromProperties_WithErrorCode testa getErrorCodeFromProperties com errorCode
func TestGetErrorCodeFromProperties_WithErrorCode(t *testing.T) {
	// Arrange
	properties := map[string]interface{}{
		"errorCode": "INVALID_CREDENTIALS",
		"timestamp": "2023-01-01T00:00:00Z",
	}

	// Act
	result := getErrorCodeFromProperties(properties)

	// Assert
	assert.Equal(t, "INVALID_CREDENTIALS", result)
}

// TestGetErrorCodeFromProperties_WithoutErrorCode testa getErrorCodeFromProperties sem errorCode
func TestGetErrorCodeFromProperties_WithoutErrorCode(t *testing.T) {
	// Arrange
	properties := map[string]interface{}{
		"timestamp": "2023-01-01T00:00:00Z",
	}

	// Act
	result := getErrorCodeFromProperties(properties)

	// Assert
	assert.Equal(t, "UNKNOWN_ERROR", result)
}

// TestGetErrorCodeFromProperties_EmptyProperties testa getErrorCodeFromProperties com propriedades vazias
func TestGetErrorCodeFromProperties_EmptyProperties(t *testing.T) {
	// Arrange
	properties := map[string]interface{}{}

	// Act
	result := getErrorCodeFromProperties(properties)

	// Assert
	assert.Equal(t, "UNKNOWN_ERROR", result)
}

// TestGetErrorCodeFromProperties_NilProperties testa getErrorCodeFromProperties com propriedades nil
func TestGetErrorCodeFromProperties_NilProperties(t *testing.T) {
	// Arrange
	var properties map[string]interface{}

	// Act
	result := getErrorCodeFromProperties(properties)

	// Assert
	assert.Equal(t, "UNKNOWN_ERROR", result)
}

// TestGetErrorCodeFromProperties_WrongType testa getErrorCodeFromProperties com tipo incorreto
func TestGetErrorCodeFromProperties_WrongType(t *testing.T) {
	// Arrange
	properties := map[string]interface{}{
		"errorCode": 123, // Deveria ser string
	}

	// Act
	result := getErrorCodeFromProperties(properties)

	// Assert
	assert.Equal(t, "UNKNOWN_ERROR", result)
}

// TestAuthClient_Login_RequestError testa login com erro de requisição
func TestAuthClient_Login_RequestError(t *testing.T) {
	// Arrange
	// Criar um servidor que não responde (timeout)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simular timeout - não responder
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	// Criar cliente com timeout muito baixo para forçar erro
	client := NewAuthClient(server.URL, 100*time.Millisecond)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	assert.Contains(t, err.Error(), "erro ao fazer requisição de login")
}

// TestAuthClient_RefreshToken_RequestError testa refresh token com erro de requisição
func TestAuthClient_RefreshToken_RequestError(t *testing.T) {
	// Arrange
	// Criar um servidor que não responde (timeout)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simular timeout - não responder
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	// Criar cliente com timeout muito baixo para forçar erro
	client := NewAuthClient(server.URL, 100*time.Millisecond)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	req := inboundDto.RefreshTokenRequestDTO{
		Token: "refresh_token",
	}
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	result, err := client.RefreshToken(ctx, req, tenantId, correlationID, "keepguard-default-client")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.RefreshTokenResponseDTO{}, result)
	assert.Contains(t, err.Error(), "erro ao fazer requisição de refresh")
}

// TestAuthClient_Logout_RequestError testa logout com erro de requisição
func TestAuthClient_Logout_RequestError(t *testing.T) {
	// Arrange
	// Criar um servidor que não responde (timeout)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simular timeout - não responder
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	// Criar cliente com timeout muito baixo para forçar erro
	client := NewAuthClient(server.URL, 100*time.Millisecond)
	ctx := authctx.WithCompanyID(context.Background(), "test-company")
	token := "access_token"
	tenantId := "test-app"
	correlationID := "test-correlation"

	// Act
	err := client.Logout(ctx, token, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao fazer requisição de logout")
}
