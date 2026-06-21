package auth

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserClient é um mock para UserClient
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetByEmail(ctx context.Context, email, xApplication, correlationID string) (outboundDto.UserByEmailResponseDTO, error) {
	args := m.Called(ctx, email, xApplication, correlationID)
	return args.Get(0).(outboundDto.UserByEmailResponseDTO), args.Error(1)
}

func TestResetPasswordUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, logger)

	ctx := context.Background()
	email := "user@example.com"
	resetToken := "valid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedReq := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "code-user-123",
		ResetToken:         resetToken,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
		MessageType:        "EMAIL",
		TemplateType:       "RECUPERACAO_SENHA",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockUserClient.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("ResetPassword", ctx, expectedReq, xApplication, correlationID).Return(nil)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestResetPasswordUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, logger)

	ctx := context.Background()
	email := "notfound@example.com"
	resetToken := "valid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	userError := &appdto.HTTPError{
		StatusCode: 404,
		Message:    "Usuário não encontrado",
		ErrorCode:  "USER_NOT_FOUND",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockUserClient.On("GetByEmail", ctx, email, xApplication, correlationID).Return(outboundDto.UserByEmailResponseDTO{}, userError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	// AuthClient NÃO deve ser chamado se o usuário não for encontrado
	mockAuthClient.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestResetPasswordUseCase_Execute_UserNotActive(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, logger)

	ctx := context.Background()
	email := "user@example.com"
	resetToken := "valid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "BLOCKED", // Status não é ACTIVE
		EmailVerified: true,
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockUserClient.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedUser, nil)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &pkg.AppError{}, err)
	appErr := err.(*pkg.AppError)
	assert.Equal(t, 400, appErr.StatusCode)
	assert.Equal(t, "USER_NOT_ACTIVE", appErr.Code)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	// AuthClient NÃO deve ser chamado se o usuário não estiver ativo
	mockAuthClient.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestResetPasswordUseCase_Execute_InvalidResetToken(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, logger)

	ctx := context.Background()
	email := "user@example.com"
	resetToken := "invalid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedReq := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "code-user-123",
		ResetToken:         resetToken,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
		MessageType:        "EMAIL",
		TemplateType:       "RECUPERACAO_SENHA",
	}

	authError := &appdto.HTTPError{
		StatusCode: 400,
		Message:    "Token de reset inválido ou expirado",
		ErrorCode:  "INVALID_RESET_TOKEN",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockUserClient.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("ResetPassword", ctx, expectedReq, xApplication, correlationID).Return(authError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestResetPasswordUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, logger)

	ctx := context.Background()
	email := "user@example.com"
	resetToken := "valid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"
	xApplication := "invalid-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	// Configurar mock do CompanyClient para retornar erro
	companyError := &appdto.HTTPError{
		StatusCode: 404,
		Message:    "Empresa não encontrada",
		ErrorCode:  "COMPANY_NOT_FOUND",
	}

	mockCompanyClient.On("GetByXApplication", ctx, xApplication, correlationID).Return(authclient.CompanySimpleResponseDTO{}, companyError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	// UserClient e AuthClient NÃO devem ser chamados se CompanyClient falhar
	mockUserClient.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockAuthClient.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestResetPasswordUseCase_Execute_PasswordMismatch(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)

	ctx := context.Background()
	email := "user@example.com"
	resetToken := "valid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "differentpass456"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	// Validação no comando detecta a diferença
	err := command.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "newPassword e confirmNewPassword devem ser iguais")
	// Nenhum mock deve ser chamado se a validação falhar
	mockCompanyClient.AssertNotCalled(t, "GetByXApplication", mock.Anything, mock.Anything, mock.Anything)
	mockUserClient.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockAuthClient.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestResetPasswordUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela o contexto imediatamente

	email := "user@example.com"
	resetToken := "valid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedReq := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "code-user-123",
		ResetToken:         resetToken,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
		MessageType:        "EMAIL",
		TemplateType:       "RECUPERACAO_SENHA",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockUserClient.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("ResetPassword", ctx, expectedReq, xApplication, correlationID).Return(context.Canceled)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestResetPasswordUseCase_Execute_AuthServiceError(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, logger)

	ctx := context.Background()
	email := "user@example.com"
	resetToken := "valid-reset-token"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewResetPasswordCommand(
		email,
		resetToken,
		newPassword,
		confirmNewPassword,
		xApplication,
		correlationID,
		ctx,
	)

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedReq := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "code-user-123",
		ResetToken:         resetToken,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
		MessageType:        "EMAIL",
		TemplateType:       "RECUPERACAO_SENHA",
	}

	authError := errors.New("auth service internal error")

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockUserClient.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("ResetPassword", ctx, expectedReq, xApplication, correlationID).Return(authError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}
