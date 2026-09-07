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
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockDevicePort struct {
	mock.Mock
}

func (m *MockDevicePort) SendChallenge(ctx context.Context, cmd appdto.SendDeviceChallengeCommand) (map[string]any, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockDevicePort) VerifyChallenge(ctx context.Context, cmd appdto.VerifyDeviceChallengeCommand) (appdto.AuthResponseDTO, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(appdto.AuthResponseDTO), args.Error(1)
}

func (m *MockDevicePort) QuickRevoke(ctx context.Context, cmd appdto.QuickRevokeCommand) (appdto.QuickRevokeViewDTO, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(appdto.QuickRevokeViewDTO), args.Error(1)
}

type MockSessionPort struct {
	mock.Mock
}

func (m *MockSessionPort) ListMe(ctx context.Context, query appdto.ListUserSessionsQuery) ([]appdto.DeviceSessionDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]appdto.DeviceSessionDTO), args.Error(1)
}

func (m *MockSessionPort) Revoke(ctx context.Context, cmd appdto.RevokeSessionCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockSessionPort) RevokeOthers(ctx context.Context, cmd appdto.RevokeAllOtherSessionsCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockSessionPort) ListTenantUser(ctx context.Context, query appdto.ListTenantUserSessionsQuery) ([]appdto.DeviceSessionDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]appdto.DeviceSessionDTO), args.Error(1)
}

func (m *MockSessionPort) RevokeTenantUser(ctx context.Context, cmd appdto.RevokeTenantUserSessionCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockSessionPort) SearchTenant(ctx context.Context, query appdto.SearchTenantSessionsQuery) (appdto.PaginatedDeviceSessionResponseDTO, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(appdto.PaginatedDeviceSessionResponseDTO), args.Error(1)
}

type MockBlacklistPort struct {
	mock.Mock
}

func (m *MockBlacklistPort) ListMe(ctx context.Context, query appdto.ListDeviceBlacklistQuery) ([]appdto.DeviceBlacklistDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]appdto.DeviceBlacklistDTO), args.Error(1)
}

func (m *MockBlacklistPort) AddMe(ctx context.Context, cmd appdto.AddDeviceBlacklistCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockBlacklistPort) RemoveMe(ctx context.Context, cmd appdto.RemoveDeviceBlacklistCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockBlacklistPort) SearchAdmin(ctx context.Context, query appdto.SearchAdminDeviceBlacklistQuery) (appdto.PaginatedDeviceBlacklistResponseDTO, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(appdto.PaginatedDeviceBlacklistResponseDTO), args.Error(1)
}

func (m *MockBlacklistPort) AdminAdd(ctx context.Context, cmd appdto.AdminAddDeviceBlacklistCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockBlacklistPort) AdminRemove(ctx context.Context, cmd appdto.AdminRemoveDeviceBlacklistCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockBlacklistPort) ListTenantUser(ctx context.Context, query appdto.ListTenantUserBlacklistQuery) ([]appdto.AdminDeviceBlacklistEntryDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]appdto.AdminDeviceBlacklistEntryDTO), args.Error(1)
}

type MockLifecyclePort struct {
	mock.Mock
}

func (m *MockLifecyclePort) BlockMe(ctx context.Context, cmd appdto.AccountLifecycleCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockLifecyclePort) DeleteMe(ctx context.Context, cmd appdto.AccountLifecycleCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

type MockLoginUseCase struct {
	mock.Mock
}

func (m *MockLoginUseCase) Execute(ctx context.Context, command appdto.LoginCommand) (dto.AuthResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(dto.AuthResponseDTO), args.Error(1)
}

type MockRefreshUseCase struct {
	mock.Mock
}

func (m *MockRefreshUseCase) Execute(ctx context.Context, command appdto.RefreshTokenCommand) (dto.RefreshTokenResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(dto.RefreshTokenResponseDTO), args.Error(1)
}

type MockLogoutUseCase struct {
	mock.Mock
}

func (m *MockLogoutUseCase) Execute(ctx context.Context, command appdto.LogoutCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

type MockValidateTokenUseCase struct {
	mock.Mock
}

func (m *MockValidateTokenUseCase) Execute(ctx context.Context, command appdto.ValidateTokenCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

type MockChangePasswordUseCase struct {
	mock.Mock
}

func (m *MockChangePasswordUseCase) Execute(ctx context.Context, command appdto.ChangePasswordCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

type MockResetPasswordUseCase struct {
	mock.Mock
}

func (m *MockResetPasswordUseCase) Execute(ctx context.Context, command appdto.ResetPasswordCommand) error {
	args := m.Called(command)
	return args.Error(0)
}

func newAuthHandlers(
	login *MockLoginUseCase,
	refresh *MockRefreshUseCase,
	logout *MockLogoutUseCase,
	devicePort *MockDevicePort,
	sessionPort *MockSessionPort,
	blacklistPort *MockBlacklistPort,
	lifecyclePort *MockLifecyclePort,
) *AuthHandlers {
	logger, _ := zap.NewDevelopment()
	return NewAuthHandlers(
		login,
		refresh,
		logout,
		new(MockValidateTokenUseCase),
		new(MockChangePasswordUseCase),
		new(MockResetPasswordUseCase),
		devicePort,
		sessionPort,
		blacklistPort,
		lifecyclePort,
		logger,
	)
}

func setupTestHandlers() (*AuthHandlers, *MockLoginUseCase, *MockRefreshUseCase, *MockLogoutUseCase) {
	mockLoginUseCase := new(MockLoginUseCase)
	mockRefreshUseCase := new(MockRefreshUseCase)
	mockLogoutUseCase := new(MockLogoutUseCase)
	handlers := newAuthHandlers(
		mockLoginUseCase,
		mockRefreshUseCase,
		mockLogoutUseCase,
		new(MockDevicePort),
		new(MockSessionPort),
		new(MockBlacklistPort),
		new(MockLifecyclePort),
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
	mockLoginUseCase.On("Execute", mock.Anything, mock.Anything).Return(expectedResponse, nil)

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

	mockLoginUseCase.On("Execute", mock.Anything, mock.Anything).Return(dto.AuthResponseDTO{}, authError)

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

	mockRefreshUseCase.On("Execute", mock.Anything, mock.Anything).Return(expectedResponse, nil)

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

	mockLogoutUseCase.On("Execute", mock.Anything, mock.Anything).Return(nil)

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
	correlationID := GetCorrelationID(c)

	// Assert
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
	correlationID := GetCorrelationID(c)

	// Assert
	assert.NotEmpty(t, correlationID)
	assert.Equal(t, correlationID, c.Response().Header().Get("X-Correlation-ID"))
}

func testJWTWithSub(sub string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"` + sub + `"}`))
	return header + "." + payload + ".sig"
}

func setupLifecycleHandlers() (*AuthHandlers, *MockLifecyclePort) {
	mockLifecycle := new(MockLifecyclePort)
	handlers := newAuthHandlers(
		new(MockLoginUseCase),
		new(MockRefreshUseCase),
		new(MockLogoutUseCase),
		new(MockDevicePort),
		new(MockSessionPort),
		new(MockBlacklistPort),
		mockLifecycle,
	)
	return handlers, mockLifecycle
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
	handlers, mockLifecycle := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", token, `{"reason":"quero bloquear"}`)

	mockLifecycle.On("BlockMe", mock.Anything, mock.MatchedBy(func(cmd appdto.AccountLifecycleCommand) bool {
		return cmd.CodeUser == "user-sub-1" &&
			cmd.Token == token &&
			cmd.Reason == "quero bloquear" &&
			cmd.TenantID == "tenant-1" &&
			cmd.CorrelationID == "corr-1"
	})).Return(nil)

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockLifecycle.AssertExpectations(t)
}

func TestDeleteMeHandler_ForwardsJWTAndExternalID(t *testing.T) {
	handlers, mockLifecycle := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodDelete, "/api/v1/users/me", token, `{"reason":"encerrar conta"}`)

	mockLifecycle.On("DeleteMe", mock.Anything, mock.MatchedBy(func(cmd appdto.AccountLifecycleCommand) bool {
		return cmd.CodeUser == "user-sub-1" &&
			cmd.Token == token &&
			cmd.Reason == "encerrar conta"
	})).Return(nil)

	err := handlers.DeleteMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockLifecycle.AssertExpectations(t)
}

func TestBlockMeHandler_Propagates403(t *testing.T) {
	handlers, mockLifecycle := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", token, `{"reason":"quero bloquear"}`)

	mockLifecycle.On("BlockMe", mock.Anything, mock.Anything).
		Return(&appdto.HTTPError{StatusCode: http.StatusForbidden, ErrorCode: "FORBIDDEN", Message: "Sem permissão"})

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response pkg.ErrorResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "FORBIDDEN", response.Error)
	mockLifecycle.AssertExpectations(t)
}

func TestBlockMeHandler_UnauthorizedWithoutToken(t *testing.T) {
	handlers, mockLifecycle := setupLifecycleHandlers()
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", "", `{"reason":"quero bloquear"}`)

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockLifecycle.AssertNotCalled(t, "BlockMe")
}

func TestBlockMeHandler_BadRequestWithoutReason(t *testing.T) {
	handlers, mockLifecycle := setupLifecycleHandlers()
	token := testJWTWithSub("user-sub-1")
	c, rec := newLifecycleContext(http.MethodPost, "/api/v1/users/me/block", token, `{"reason":"   "}`)

	err := handlers.BlockMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockLifecycle.AssertNotCalled(t, "BlockMe")
}

func TestGetTenantAndClientId_FromHeaderWhenNoJWT(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("X-Tenant-Id", "tenant-header")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tenantId, _, err := GetTenantAndClientId(c)
	assert.NoError(t, err)
	assert.Equal(t, "tenant-header", tenantId)
}

func TestGetTenantAndClientId_PrefersJWTOverHeader(t *testing.T) {
	e := echo.New()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"u1","tenant_id":"tenant-jwt"}`))
	token := header + "." + payload + ".sig"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-Id", "tenant-jwt")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tenantId, _, err := GetTenantAndClientId(c)
	assert.NoError(t, err)
	assert.Equal(t, "tenant-jwt", tenantId)
}

func TestGetTenantAndClientId_FromJWTWithoutHeader(t *testing.T) {
	e := echo.New()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"u1","tenant_id":"tenant-jwt"}`))
	token := header + "." + payload + ".sig"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tenantId, _, err := GetTenantAndClientId(c)
	assert.NoError(t, err)
	assert.Equal(t, "tenant-jwt", tenantId)
}

func TestGetTenantAndClientId_MismatchHeaderAndJWT(t *testing.T) {
	e := echo.New()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"u1","tenant_id":"tenant-jwt"}`))
	token := header + "." + payload + ".sig"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-Id", "other-tenant")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, _, err := GetTenantAndClientId(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "não corresponde")
}

func TestGetTenantAndClientId_MissingWhenNoHeaderAndNoJWT(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, _, err := GetTenantAndClientId(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "X-Tenant-Id")
}

func TestSendDeviceChallengeHandler_Success(t *testing.T) {
	devicePort := new(MockDevicePort)
	handlers := newAuthHandlers(
		new(MockLoginUseCase),
		new(MockRefreshUseCase),
		new(MockLogoutUseCase),
		devicePort,
		new(MockSessionPort),
		new(MockBlacklistPort),
		new(MockLifecyclePort),
	)

	e := echo.New()
	reqBody, _ := json.Marshal(dto.DeviceChallengeSendRequestDTO{ChallengeSessionID: "sess-1", Channel: "email"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/challenge/send", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	devicePort.On("SendChallenge", mock.Anything, mock.MatchedBy(func(cmd appdto.SendDeviceChallengeCommand) bool {
		return cmd.ChallengeSessionID == "sess-1" && cmd.Channel == "email" && cmd.TenantID == "tenant-1"
	})).Return(map[string]any{"status": "sent"}, nil)

	err := handlers.SendDeviceChallengeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	devicePort.AssertExpectations(t)
}

func TestQuickRevokeHandler_JSONAndHTML(t *testing.T) {
	devicePort := new(MockDevicePort)
	handlers := newAuthHandlers(
		new(MockLoginUseCase),
		new(MockRefreshUseCase),
		new(MockLogoutUseCase),
		devicePort,
		new(MockSessionPort),
		new(MockBlacklistPort),
		new(MockLifecyclePort),
	)

	view := appdto.QuickRevokeViewDTO{
		Message: "Sessão revogada.",
		Payload: map[string]any{"message": "Sessão revogada.", "ok": true},
	}

	t.Run("json", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/device/quick-revoke?token=abc", nil)
		req.Header.Set("X-Tenant-Id", "tenant-1")
		req.Header.Set("X-Correlation-ID", "corr-1")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		devicePort.On("QuickRevoke", mock.Anything, mock.MatchedBy(func(cmd appdto.QuickRevokeCommand) bool {
			return cmd.Token == "abc" && cmd.Blacklist && cmd.TenantID == "tenant-1"
		})).Return(view, nil).Once()

		err := handlers.QuickRevokeHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"ok":true`)
	})

	t.Run("html", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/device/quick-revoke?token=abc", nil)
		req.Header.Set("Accept", "text/html")
		req.Header.Set("X-Correlation-ID", "corr-1")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		devicePort.On("QuickRevoke", mock.Anything, mock.MatchedBy(func(cmd appdto.QuickRevokeCommand) bool {
			return cmd.Token == "abc" && cmd.TenantID == "keepguard-default"
		})).Return(view, nil).Once()

		err := handlers.QuickRevokeHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Sessão revogada.")
		assert.Contains(t, rec.Body.String(), "Acesso Revogado com Sucesso")
	})
}

func TestListUserSessionsHandler_Success(t *testing.T) {
	sessionPort := new(MockSessionPort)
	handlers := newAuthHandlers(
		new(MockLoginUseCase),
		new(MockRefreshUseCase),
		new(MockLogoutUseCase),
		new(MockDevicePort),
		sessionPort,
		new(MockBlacklistPort),
		new(MockLifecyclePort),
	)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/sessions", nil)
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("X-Device-Id", "dev-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sessionPort.On("ListMe", mock.Anything, mock.MatchedBy(func(q appdto.ListUserSessionsQuery) bool {
		return q.Token == "tok" && q.DeviceID == "dev-1" && q.TenantID == "tenant-1"
	})).Return([]appdto.DeviceSessionDTO{{DeviceID: "dev-1", IsCurrent: true}}, nil)

	err := handlers.ListUserSessionsHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	sessionPort.AssertExpectations(t)
}

func TestListDeviceBlacklistHandler_Success(t *testing.T) {
	blacklistPort := new(MockBlacklistPort)
	handlers := newAuthHandlers(
		new(MockLoginUseCase),
		new(MockRefreshUseCase),
		new(MockLogoutUseCase),
		new(MockDevicePort),
		new(MockSessionPort),
		blacklistPort,
		new(MockLifecyclePort),
	)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/devices/blacklist", nil)
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	blacklistPort.On("ListMe", mock.Anything, mock.MatchedBy(func(q appdto.ListDeviceBlacklistQuery) bool {
		return q.Token == "tok" && q.TenantID == "tenant-1"
	})).Return([]appdto.DeviceBlacklistDTO{{DeviceID: "dev-blocked"}}, nil)

	err := handlers.ListDeviceBlacklistHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	blacklistPort.AssertExpectations(t)
}
