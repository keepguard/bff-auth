package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAuthClient para os testes de handlers
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Login(ctx context.Context, req dto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (dto.AuthResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(dto.AuthResponseDTO), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, req dto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (dto.RefreshTokenResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(dto.RefreshTokenResponseDTO), args.Error(1)
}

func (m *MockAuthClient) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	args := m.Called(ctx, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	args := m.Called(ctx, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	args := m.Called(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
	return args.Error(0)
}

func (m *MockAuthClient) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	args := m.Called(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
	return args.Error(0)
}

func (m *MockAuthClient) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(outboundDto.GenerateResetTokenMSResponseDTO), args.Error(1)
}

func (m *MockAuthClient) SendDeviceChallenge(ctx context.Context, req dto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) VerifyDeviceChallenge(ctx context.Context, req dto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (dto.AuthResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(dto.AuthResponseDTO), args.Error(1)
}

func (m *MockAuthClient) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]dto.DeviceSessionDTO, error) {
	args := m.Called(ctx, token, deviceId, tenantId, correlationID)
	return args.Get(0).([]dto.DeviceSessionDTO), args.Error(1)
}

func (m *MockAuthClient) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	args := m.Called(ctx, deviceIdToRevoke, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	args := m.Called(ctx, token, currentDeviceId, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error) {
	args := m.Called(ctx, token, blacklist, tenantId, correlationID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]dto.DeviceBlacklistDTO, error) {
	args := m.Called(ctx, token, tenantId, correlationID)
	return args.Get(0).([]dto.DeviceBlacklistDTO), args.Error(1)
}

func (m *MockAuthClient) AddDeviceToBlacklist(ctx context.Context, req dto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	args := m.Called(ctx, req, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	args := m.Called(ctx, deviceId, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (dto.PaginatedDeviceBlacklistResponseDTO, error) {
	args := m.Called(ctx, queryParams, token, tenantId, correlationID)
	return args.Get(0).(dto.PaginatedDeviceBlacklistResponseDTO), args.Error(1)
}

func (m *MockAuthClient) AdminAddDeviceToBlacklist(ctx context.Context, req dto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	args := m.Called(ctx, req, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error {
	args := m.Called(ctx, deviceId, userId, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (outboundDto.UserByCodeResponseDTO, error) {
	args := m.Called(ctx, codeUser, token, tenantId, correlationID)
	return args.Get(0).(outboundDto.UserByCodeResponseDTO), args.Error(1)
}

type MockCompanyClient struct {
	mock.Mock
}

func (m *MockCompanyClient) GetByTenantId(ctx context.Context, tenantId, correlationID string) (authclient.CompanySimpleResponseDTO, error) {
	args := m.Called(ctx, tenantId, correlationID)
	return args.Get(0).(authclient.CompanySimpleResponseDTO), args.Error(1)
}

func (m *MockAuthClient) BlockUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	args := m.Called(ctx, idUserExternal, reason, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) DeleteUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	args := m.Called(ctx, idUserExternal, reason, token, tenantId, correlationID)
	return args.Error(0)
}

// MockUseCases são mocks para os casos de uso
type MockLoginUseCase struct {
	mock.Mock
}

func (m *MockLoginUseCase) Execute(command appdto.LoginCommand) (dto.AuthResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(dto.AuthResponseDTO), args.Error(1)
}

type MockRefreshUseCase struct {
	mock.Mock
}

func (m *MockRefreshUseCase) Execute(command appdto.RefreshTokenCommand) (dto.RefreshTokenResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(dto.RefreshTokenResponseDTO), args.Error(1)
}

type MockLogoutUseCase struct {
	mock.Mock
}

func (m *MockLogoutUseCase) Execute(command appdto.LogoutCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

type MockValidateTokenUseCase struct {
	mock.Mock
}

func (m *MockValidateTokenUseCase) Execute(command appdto.ValidateTokenCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

type MockChangePasswordUseCase struct {
	mock.Mock
}

func (m *MockChangePasswordUseCase) Execute(command appdto.ChangePasswordCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

type MockResetPasswordUseCase struct {
	mock.Mock
}

func (m *MockResetPasswordUseCase) Execute(command appdto.ResetPasswordCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

func setupTestHandlers() (*AuthHandlers, *MockLoginUseCase, *MockRefreshUseCase, *MockLogoutUseCase) {
	mockLoginUseCase := new(MockLoginUseCase)
	mockRefreshUseCase := new(MockRefreshUseCase)
	mockLogoutUseCase := new(MockLogoutUseCase)
	mockValidateTokenUseCase := new(MockValidateTokenUseCase)
	mockChangePasswordUseCase := new(MockChangePasswordUseCase)
	mockResetPasswordUseCase := new(MockResetPasswordUseCase)
	mockAuthClient := new(MockAuthClient)
	logger, _ := zap.NewDevelopment()

	handlers := NewAuthHandlers(
		mockLoginUseCase,
		mockRefreshUseCase,
		mockLogoutUseCase,
		mockValidateTokenUseCase,
		mockChangePasswordUseCase,
		mockResetPasswordUseCase,
		mockAuthClient,
		new(MockCompanyClient),
		logger,
	)

	return handlers, mockLoginUseCase, mockRefreshUseCase, mockLogoutUseCase
}

func TestAuthHandlers_LoginHandler_Success(t *testing.T) {
	// Arrange
	handlers, mockLoginUseCase, _, _ := setupTestHandlers()

	e := echo.New()
	reqBody := dto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedResponse := dto.AuthResponseDTO{
		Token:     "access_token",
		ExpiresIn: 3600,
	}

	// O teste precisa ser ajustado para usar o novo command
	// Por enquanto, vamos usar mock.Anything para o comando
	mockLoginUseCase.On("Execute", mock.Anything).Return(expectedResponse, nil)

	// Act
	err := handlers.LoginHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.AuthResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	mockLoginUseCase.AssertExpectations(t)
}

func TestAuthHandlers_LoginHandler_InvalidJSON(t *testing.T) {
	// Arrange
	handlers, _, _, _ := setupTestHandlers()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.LoginHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", response.Error)
	assert.Equal(t, "Requisição inválida", response.Message)

}

func TestAuthHandlers_LoginHandler_UseCaseError(t *testing.T) {
	// Arrange
	handlers, mockLoginUseCase, _, _ := setupTestHandlers()

	e := echo.New()
	reqBody := dto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	authError := &appdto.HTTPError{
		StatusCode: 401,
		Message:    "Credenciais inválidas",
		ErrorCode:  "INVALID_CREDENTIALS",
	}

	mockLoginUseCase.On("Execute", mock.Anything).Return(dto.AuthResponseDTO{}, authError)

	// Act
	err := handlers.LoginHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", response.Error)
	assert.Equal(t, "Credenciais inválidas", response.Message)

	mockLoginUseCase.AssertExpectations(t)
}

func TestAuthHandlers_RefreshHandler_Success(t *testing.T) {
	// Arrange
	handlers, _, mockRefreshUseCase, _ := setupTestHandlers()

	e := echo.New()
	reqBody := dto.RefreshTokenRequestDTO{
		Token: "valid_refresh_token",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedResponse := dto.RefreshTokenResponseDTO{
		Token:     "new_access_token",
		ExpiresIn: 3600,
	}

	mockRefreshUseCase.On("Execute", mock.Anything).Return(expectedResponse, nil)

	// Act
	err := handlers.RefreshHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.RefreshTokenResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	mockRefreshUseCase.AssertExpectations(t)
}

func TestAuthHandlers_LogoutHandler_Success(t *testing.T) {
	// Arrange
	handlers, _, _, mockLogoutUseCase := setupTestHandlers()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLogoutUseCase.On("Execute", mock.Anything).Return(nil)

	// Act
	err := handlers.LogoutHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Logout realizado com sucesso", response["message"])

	mockLogoutUseCase.AssertExpectations(t)
}

func TestAuthHandlers_LogoutHandler_NoAuthorizationHeader(t *testing.T) {
	// Arrange
	handlers, _, _, _ := setupTestHandlers()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.LogoutHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", response.Error)
	assert.Equal(t, "Token de autorização não fornecido", response.Message)
}

func TestAuthHandlers_LogoutHandler_InvalidToken(t *testing.T) {
	// Arrange
	handlers, _, _, _ := setupTestHandlers()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer ")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	req.Header.Set("X-Tenant-Id", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handlers.LogoutHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", response.Error)
	assert.Equal(t, "Token inválido", response.Message)
}

func TestGetOrCreateCorrelationID_WithExistingID(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "existing-correlation-id")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	correlationID, err := GetCorrelationID(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "existing-correlation-id", correlationID)
	assert.Equal(t, "existing-correlation-id", c.Response().Header().Get("X-Correlation-ID"))
}

func TestGetCorrelationID_WithoutExistingID(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	correlationID, err := GetCorrelationID(c)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, correlationID)
	assert.Equal(t, "Header X-Correlation-ID é obrigatório", err.Error())
}

func testJWTWithSub(sub string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"` + sub + `"}`))
	return header + "." + payload + ".sig"
}

func setupLifecycleHandlers() (*AuthHandlers, *MockAuthClient) {
	mockLoginUseCase := new(MockLoginUseCase)
	mockRefreshUseCase := new(MockRefreshUseCase)
	mockLogoutUseCase := new(MockLogoutUseCase)
	mockValidateTokenUseCase := new(MockValidateTokenUseCase)
	mockChangePasswordUseCase := new(MockChangePasswordUseCase)
	mockResetPasswordUseCase := new(MockResetPasswordUseCase)
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	mockCompanyClient.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(authclient.CompanySimpleResponseDTO{ID: "company-1"}, nil).Maybe()
	logger, _ := zap.NewDevelopment()

	handlers := NewAuthHandlers(
		mockLoginUseCase,
		mockRefreshUseCase,
		mockLogoutUseCase,
		mockValidateTokenUseCase,
		mockChangePasswordUseCase,
		mockResetPasswordUseCase,
		mockAuthClient,
		mockCompanyClient,
		logger,
	)
	return handlers, mockAuthClient
}

func newLifecycleContext(method, path, token, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestBlockMeHandler_ForwardsJWTAndExternalID(t *testing.T) {
	handlers, mockAuthClient := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", token, `{"reason":"quero bloquear"}`)

	mockAuthClient.On("GetUserByCodeUser", mock.Anything, "user-sub-1", token, "tenant-1", "corr-1").
		Return(outboundDto.UserByCodeResponseDTO{IDUserExternal: "ext-id-1"}, nil)
	mockAuthClient.On("BlockUser", mock.Anything, "ext-id-1", "quero bloquear", token, "tenant-1", "corr-1").
		Return(nil)

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestDeleteMeHandler_ForwardsJWTAndExternalID(t *testing.T) {
	handlers, mockAuthClient := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodDelete, "/api/v1/users/me", token, `{"reason":"encerrar conta"}`)

	mockAuthClient.On("GetUserByCodeUser", mock.Anything, "user-sub-1", token, "tenant-1", "corr-1").
		Return(outboundDto.UserByCodeResponseDTO{IDUserExternal: "ext-id-1"}, nil)
	mockAuthClient.On("DeleteUser", mock.Anything, "ext-id-1", "encerrar conta", token, "tenant-1", "corr-1").
		Return(nil)

	err := handlers.DeleteMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestBlockMeHandler_Propagates403(t *testing.T) {
	handlers, mockAuthClient := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", token, `{"reason":"quero bloquear"}`)

	mockAuthClient.On("GetUserByCodeUser", mock.Anything, "user-sub-1", token, "tenant-1", "corr-1").
		Return(outboundDto.UserByCodeResponseDTO{IDUserExternal: "ext-id-1"}, nil)
	mockAuthClient.On("BlockUser", mock.Anything, "ext-id-1", "quero bloquear", token, "tenant-1", "corr-1").
		Return(&appdto.HTTPError{StatusCode: http.StatusForbidden, ErrorCode: "FORBIDDEN", Message: "Sem permissão"})

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response pkg.ErrorResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "FORBIDDEN", response.Error)
	mockAuthClient.AssertExpectations(t)
}

func TestBlockMeHandler_UnauthorizedWithoutToken(t *testing.T) {
	handlers, mockAuthClient := setupLifecycleHandlers()
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", "", `{"reason":"quero bloquear"}`)

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockAuthClient.AssertNotCalled(t, "BlockUser")
}

func TestBlockMeHandler_BadRequestWithoutReason(t *testing.T) {
	handlers, mockAuthClient := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", token, `{"reason":"   "}`)

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthClient.AssertNotCalled(t, "GetUserByCodeUser")
	mockAuthClient.AssertNotCalled(t, "BlockUser")
}
