package auth

import (
	"context"
	"errors"
	"testing"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/keepguard/bff-auth/internal/infrastructure/resilience"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthClient é um mock para AuthClient
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(inboundDto.AuthResponseDTO), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(inboundDto.RefreshTokenResponseDTO), args.Error(1)
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

func (m *MockAuthClient) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(inboundDto.AuthResponseDTO), args.Error(1)
}

func (m *MockAuthClient) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	args := m.Called(ctx, token, deviceId, tenantId, correlationID)
	return args.Get(0).([]inboundDto.DeviceSessionDTO), args.Error(1)
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

func (m *MockAuthClient) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error) {
	args := m.Called(ctx, token, tenantId, correlationID)
	return args.Get(0).([]inboundDto.DeviceBlacklistDTO), args.Error(1)
}

func (m *MockAuthClient) AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	args := m.Called(ctx, req, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	args := m.Called(ctx, deviceId, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceBlacklistResponseDTO, error) {
	args := m.Called(ctx, queryParams, token, tenantId, correlationID)
	return args.Get(0).(inboundDto.PaginatedDeviceBlacklistResponseDTO), args.Error(1)
}

func (m *MockAuthClient) AdminAddDeviceToBlacklist(ctx context.Context, req inboundDto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	args := m.Called(ctx, req, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error {
	args := m.Called(ctx, deviceId, userId, token, tenantId, correlationID)
	return args.Error(0)
}

// MockMetricsRecorder é um mock para MetricsRecorder
type MockMetricsRecorder struct {
	mock.Mock
}

func (m *MockMetricsRecorder) SetCircuitBreakerState(service string, state int) {
	m.Called(service, state)
}

func TestNewCircuitBreakerDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	// Act
	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &circuitBreakerDecorator{}, decorator)
}

func TestCircuitBreakerDecorator_Login_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := inboundDto.AuthResponseDTO{
		Token:     "access_token",
		ExpiresIn: 3600,
	}

	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_Login_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("auth service error")

	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, expectedError)

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_RefreshToken_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := inboundDto.RefreshTokenRequestDTO{
		Token: "valid_refresh_token",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := inboundDto.RefreshTokenResponseDTO{
		Token:     "new_access_token",
		ExpiresIn: 3600,
	}

	mockInner.On("RefreshToken", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.RefreshToken(ctx, req, tenantId, correlationID, "keepguard-default-client")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_RefreshToken_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := inboundDto.RefreshTokenRequestDTO{
		Token: "invalid_refresh_token",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("refresh token error")

	mockInner.On("RefreshToken", ctx, req, tenantId, correlationID).Return(inboundDto.RefreshTokenResponseDTO{}, expectedError)

	// Act
	result, err := decorator.RefreshToken(ctx, req, tenantId, correlationID, "keepguard-default-client")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, inboundDto.RefreshTokenResponseDTO{}, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_Logout_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	token := "valid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("Logout", ctx, token, tenantId, correlationID).Return(nil)

	// Act
	err := decorator.Logout(ctx, token, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_Logout_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	token := "invalid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("logout error")

	mockInner.On("Logout", ctx, token, tenantId, correlationID).Return(expectedError)

	// Act
	err := decorator.Logout(ctx, token, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)
}
