package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockSendResetPasswordMessageUseCase é um mock para SendResetPasswordMessageUseCase
type MockSendResetPasswordMessageUseCase struct {
	mock.Mock
}

func (m *MockSendResetPasswordMessageUseCase) Execute(command appdto.SendResetPasswordMessageCommand) (inboundDto.SendResetPasswordMessageResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(inboundDto.SendResetPasswordMessageResponseDTO), args.Error(1)
}

// setupTestMessageHandlers configura handlers para testes
func setupTestMessageHandlers() (*MessageHandlers, *MockSendResetPasswordMessageUseCase) {
	mockSendResetPasswordMessageUseCase := new(MockSendResetPasswordMessageUseCase)
	logger, _ := zap.NewDevelopment()

	handlers := NewMessageHandlers(
		mockSendResetPasswordMessageUseCase,
		logger,
	)

	return handlers, mockSendResetPasswordMessageUseCase
}

// TestMessageHandlers_SendResetPasswordMessageHandler_Success testa handler com sucesso
func TestMessageHandlers_SendResetPasswordMessageHandler_Success(t *testing.T) {
	// Arrange
	handlers, mockUseCase := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "test@example.com",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedResponse := inboundDto.SendResetPasswordMessageResponseDTO{
		Success: true,
		Message: "Mensagem enviada com sucesso",
	}

	mockUseCase.On("Execute", mock.Anything).Return(expectedResponse, nil)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response inboundDto.SendResetPasswordMessageResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	mockUseCase.AssertExpectations(t)
}

// TestMessageHandlers_SendResetPasswordMessageHandler_MissingCorrelationID testa sem correlation ID
func TestMessageHandlers_SendResetPasswordMessageHandler_MissingCorrelationID(t *testing.T) {
	// Arrange
	handlers, _ := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "test@example.com",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	// Não definir X-Correlation-ID
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_HEADER", response.Error)
	assert.Contains(t, response.Message, "X-Correlation-ID")
}

// TestMessageHandlers_SendResetPasswordMessageHandler_MissingTenantId testa sem X-Tenant-Id
func TestMessageHandlers_SendResetPasswordMessageHandler_MissingTenantId(t *testing.T) {
	// Arrange
	handlers, _ := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "test@example.com",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	// Não definir X-Tenant-Id
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_HEADER", response.Error)
	assert.Contains(t, response.Message, "X-Tenant-Id")
}

// TestMessageHandlers_SendResetPasswordMessageHandler_InvalidJSON testa JSON inválido
func TestMessageHandlers_SendResetPasswordMessageHandler_InvalidJSON(t *testing.T) {
	// Arrange
	handlers, _ := setupTestMessageHandlers()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", response.Error)
	assert.Equal(t, "Requisição inválida", response.Message)
}

// TestMessageHandlers_SendResetPasswordMessageHandler_ValidationError testa erro de validação do comando
func TestMessageHandlers_SendResetPasswordMessageHandler_ValidationError(t *testing.T) {
	// Arrange
	handlers, _ := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "", // Email vazio deve falhar na validação
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Error)
	assert.Contains(t, response.Message, "email")
}

// TestMessageHandlers_SendResetPasswordMessageHandler_UseCaseHTTPError testa erro HTTP do use case
func TestMessageHandlers_SendResetPasswordMessageHandler_UseCaseHTTPError(t *testing.T) {
	// Arrange
	handlers, mockUseCase := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "test@example.com",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	httpError := &appdto.HTTPError{
		StatusCode: 404,
		Message:    "Usuário não encontrado",
		ErrorCode:  "USER_NOT_FOUND",
	}

	mockUseCase.On("Execute", mock.Anything).Return(inboundDto.SendResetPasswordMessageResponseDTO{}, httpError)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "USER_NOT_FOUND", response.Error)
	assert.Equal(t, "Usuário não encontrado", response.Message)

	mockUseCase.AssertExpectations(t)
}

// TestMessageHandlers_SendResetPasswordMessageHandler_UseCaseAppError testa AppError do use case
func TestMessageHandlers_SendResetPasswordMessageHandler_UseCaseAppError(t *testing.T) {
	// Arrange
	handlers, mockUseCase := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "test@example.com",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	appError := &pkg.AppError{
		StatusCode: 400,
		Code:       "USER_NOT_ACTIVE",
		Message:    "Usuário não está ativo",
	}

	mockUseCase.On("Execute", mock.Anything).Return(inboundDto.SendResetPasswordMessageResponseDTO{}, appError)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "USER_NOT_ACTIVE", response.Error)
	assert.Equal(t, "Usuário não está ativo", response.Message)

	mockUseCase.AssertExpectations(t)
}

// TestMessageHandlers_SendResetPasswordMessageHandler_GenericError testa erro genérico
func TestMessageHandlers_SendResetPasswordMessageHandler_GenericError(t *testing.T) {
	// Arrange
	handlers, mockUseCase := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "test@example.com",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	genericError := assert.AnError

	mockUseCase.On("Execute", mock.Anything).Return(inboundDto.SendResetPasswordMessageResponseDTO{}, genericError)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", response.Error)
	assert.Equal(t, "Erro interno do servidor", response.Message)

	mockUseCase.AssertExpectations(t)
}

// TestMessageHandlers_SendResetPasswordMessageHandler_InvalidEmailFormat testa email com formato inválido
func TestMessageHandlers_SendResetPasswordMessageHandler_InvalidEmailFormat(t *testing.T) {
	// Arrange
	handlers, _ := setupTestMessageHandlers()

	e := echo.New()
	reqBody := inboundDto.SendResetPasswordMessageRequestDTO{
		Email: "invalid-email", // Email com formato inválido
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.SendResetPasswordMessageHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Error)
	assert.Contains(t, response.Message, "email")
}

// TestNewMessageHandlers testa a criação do handler
func TestNewMessageHandlers(t *testing.T) {
	// Arrange
	mockUseCase := new(MockSendResetPasswordMessageUseCase)
	logger, _ := zap.NewDevelopment()

	// Act
	handlers := NewMessageHandlers(mockUseCase, logger)

	// Assert
	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.sendResetPasswordMessageUseCase)
	assert.NotNil(t, handlers.logger)
}

// TestNewMessageHandlersWithLogger testa a criação do handler com interface logger
func TestNewMessageHandlersWithLogger(t *testing.T) {
	// Arrange
	mockUseCase := new(MockSendResetPasswordMessageUseCase)

	// Act
	handlers := NewMessageHandlersWithLogger(mockUseCase, nil)

	// Assert
	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.sendResetPasswordMessageUseCase)
	assert.NotNil(t, handlers.logger)
}
