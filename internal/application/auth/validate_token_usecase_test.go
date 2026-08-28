package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockLoginTokenChecker struct {
	mock.Mock
}

func (m *MockLoginTokenChecker) Check(ctx context.Context, token string) (authclient.LoginTokenPresence, string, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(authclient.LoginTokenPresence), args.String(1), args.Error(2)
}

func presentTokenChecker(token string) *MockLoginTokenChecker {
	checker := new(MockLoginTokenChecker)
	checker.On("Check", mock.Anything, token).Return(authclient.LoginTokenPresent, "code-user-1", nil)
	return checker
}

func newValidateUseCase(authClient *MockAuthClient, companyClient *MockCompanyClient, checker authclient.LoginTokenChecker) ValidateTokenUseCase {
	logger, _ := zap.NewDevelopment()
	return NewValidateTokenUseCase(authClient, companyClient, checker, logger)
}

func TestValidateTokenUseCase_Execute_Success(t *testing.T) {
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	token := "valid_token"
	checker := presentTokenChecker(token)
	useCase := newValidateUseCase(mockAuthClient, mockCompanyClient, checker)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	command := appdto.NewValidateTokenCommand(token, tenantId, correlationID, ctx)

	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil)

	err := useCase.Execute(command)

	assert.NoError(t, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	checker.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_InvalidToken(t *testing.T) {
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	token := "invalid_token"
	checker := presentTokenChecker(token)
	useCase := newValidateUseCase(mockAuthClient, mockCompanyClient, checker)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	command := appdto.NewValidateTokenCommand(token, tenantId, correlationID, ctx)

	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, tenantId, correlationID).Return(assert.AnError)

	err := useCase.Execute(command)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_RevokedInRedis(t *testing.T) {
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	token := "revoked_token"
	checker := new(MockLoginTokenChecker)
	checker.On("Check", mock.Anything, token).Return(authclient.LoginTokenAbsent, "code-user-1", nil)
	useCase := newValidateUseCase(mockAuthClient, mockCompanyClient, checker)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	command := appdto.NewValidateTokenCommand(token, tenantId, correlationID, ctx)

	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)

	err := useCase.Execute(command)

	assert.Error(t, err)
	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.StatusCode)
	assert.Equal(t, "TOKEN_REVOKED", httpErr.ErrorCode)
	mockAuthClient.AssertNotCalled(t, "ValidateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	checker.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_RedisErrorFallsBackToMSAuth(t *testing.T) {
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	token := "valid_token"
	checker := new(MockLoginTokenChecker)
	checker.On("Check", mock.Anything, token).Return(authclient.LoginTokenUnknown, "code-user-1", errors.New("redis down"))
	useCase := newValidateUseCase(mockAuthClient, mockCompanyClient, checker)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	command := appdto.NewValidateTokenCommand(token, tenantId, correlationID, ctx)

	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil)

	err := useCase.Execute(command)

	assert.NoError(t, err)
	mockAuthClient.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_CompanyNotFound(t *testing.T) {
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	checker := new(MockLoginTokenChecker)
	useCase := NewValidateTokenUseCase(mockAuthClient, mockCompanyClient, checker, logger)

	ctx := context.Background()
	token := "valid_token"
	tenantId := "invalid-app-id"
	correlationID := "test-correlation-id"
	command := appdto.NewValidateTokenCommand(token, tenantId, correlationID, ctx)

	companyError := &appdto.HTTPError{
		StatusCode: 404,
		Message:    "Empresa não encontrada",
		ErrorCode:  "COMPANY_NOT_FOUND",
	}
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(authclient.CompanySimpleResponseDTO{}, companyError)

	err := useCase.Execute(command)

	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertNotCalled(t, "ValidateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	checker.AssertNotCalled(t, "Check", mock.Anything, mock.Anything)
}

func TestValidateTokenUseCase_Execute_EmptyToken(t *testing.T) {
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	token := ""
	checker := presentTokenChecker(token)
	useCase := newValidateUseCase(mockAuthClient, mockCompanyClient, checker)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	command := appdto.NewValidateTokenCommand(token, tenantId, correlationID, ctx)

	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil)

	err := useCase.Execute(command)

	assert.NoError(t, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_ContextCancelled(t *testing.T) {
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	token := "valid_token"
	checker := presentTokenChecker(token)
	useCase := newValidateUseCase(mockAuthClient, mockCompanyClient, checker)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	command := appdto.NewValidateTokenCommand(token, tenantId, correlationID, ctx)

	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, tenantId, correlationID).Return(context.Canceled)

	err := useCase.Execute(command)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}
