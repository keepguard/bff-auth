package auth

import (
	"context"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestChangePasswordUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewChangePasswordUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	// Token JWT válido com payload contendo codeUser
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlVXNlciI6InVzZXItMTIzIiwic3ViIjoidXNlci0xMjMiLCJ1c2VybmFtZSI6InRlc3R1c2VyIn0.test"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	currentPassword := "currentpass123"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"

	command := appdto.NewChangePasswordCommand(
		token,
		currentPassword,
		newPassword,
		confirmNewPassword,
		tenantId,
		correlationID,
		"",
		"",
		"",
		"",
		"",
		ctx,
	)

	expectedReq := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-123",
		CurrentPassword:    currentPassword,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ChangePassword", ctx, expectedReq, tenantId, correlationID, "", "", "", "", "").Return(nil)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestChangePasswordUseCase_Execute_InvalidToken(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewChangePasswordUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	token := "invalid_token_format"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	currentPassword := "currentpass123"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"

	command := appdto.NewChangePasswordCommand(
		token,
		currentPassword,
		newPassword,
		confirmNewPassword,
		tenantId,
		correlationID,
		"",
		"",
		"",
		"",
		"",
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	mockCompanyClient.AssertExpectations(t)
	// AuthClient NÃO deve ser chamado se o token for inválido
	mockAuthClient.AssertNotCalled(t, "ChangePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestChangePasswordUseCase_Execute_AuthServiceError(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewChangePasswordUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlVXNlciI6InVzZXItMTIzIiwic3ViIjoidXNlci0xMjMiLCJ1c2VybmFtZSI6InRlc3R1c2VyIn0.test"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	currentPassword := "wrongcurrentpass"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"

	command := appdto.NewChangePasswordCommand(
		token,
		currentPassword,
		newPassword,
		confirmNewPassword,
		tenantId,
		correlationID,
		"",
		"",
		"",
		"",
		"",
		ctx,
	)

	expectedReq := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-123",
		CurrentPassword:    currentPassword,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
	}

	authError := &appdto.HTTPError{
		StatusCode: 400,
		Message:    "Senha atual incorreta",
		ErrorCode:  "INVALID_PASSWORD",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ChangePassword", ctx, expectedReq, tenantId, correlationID, "", "", "", "", "").Return(authError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestChangePasswordUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewChangePasswordUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlVXNlciI6InVzZXItMTIzIiwic3ViIjoidXNlci0xMjMiLCJ1c2VybmFtZSI6InRlc3R1c2VyIn0.test"
	tenantId := "invalid-app-id"
	correlationID := "test-correlation-id"
	currentPassword := "currentpass123"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"

	command := appdto.NewChangePasswordCommand(
		token,
		currentPassword,
		newPassword,
		confirmNewPassword,
		tenantId,
		correlationID,
		"",
		"",
		"",
		"",
		"",
		ctx,
	)

	// Configurar mock do CompanyClient para retornar erro
	companyError := &appdto.HTTPError{
		StatusCode: 404,
		Message:    "Empresa não encontrada",
		ErrorCode:  "COMPANY_NOT_FOUND",
	}

	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(authclient.CompanySimpleResponseDTO{}, companyError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	// AuthClient NÃO deve ser chamado se CompanyClient falhar
	mockAuthClient.AssertNotCalled(t, "ChangePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestChangePasswordUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewChangePasswordUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela o contexto imediatamente

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlVXNlciI6InVzZXItMTIzIiwic3ViIjoidXNlci0xMjMiLCJ1c2VybmFtZSI6InRlc3R1c2VyIn0.test"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	currentPassword := "currentpass123"
	newPassword := "newpass123"
	confirmNewPassword := "newpass123"

	command := appdto.NewChangePasswordCommand(
		token,
		currentPassword,
		newPassword,
		confirmNewPassword,
		tenantId,
		correlationID,
		"",
		"",
		"",
		"",
		"",
		ctx,
	)

	expectedReq := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-123",
		CurrentPassword:    currentPassword,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("ChangePassword", ctx, expectedReq, tenantId, correlationID, "", "", "", "", "").Return(context.Canceled)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestChangePasswordUseCase_Execute_PasswordMismatch(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)

	ctx := context.Background()
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlVXNlciI6InVzZXItMTIzIiwic3ViIjoidXNlci0xMjMiLCJ1c2VybmFtZSI6InRlc3R1c2VyIn0.test"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"
	currentPassword := "currentpass123"
	newPassword := "newpass123"
	confirmNewPassword := "differentpass456"

	command := appdto.NewChangePasswordCommand(
		token,
		currentPassword,
		newPassword,
		confirmNewPassword,
		tenantId,
		correlationID,
		"",
		"",
		"",
		"",
		"",
		ctx,
	)

	// Validação no comando detecta a diferença
	err := command.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Nova senha e confirmação devem ser iguais")
	// CompanyClient e AuthClient NÃO devem ser chamados se a validação falhar
	mockCompanyClient.AssertNotCalled(t, "GetByTenantId", mock.Anything, mock.Anything, mock.Anything)
	mockAuthClient.AssertNotCalled(t, "ChangePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
