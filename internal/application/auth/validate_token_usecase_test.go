package auth

import (
	"context"
	"testing"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestValidateTokenUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewValidateTokenUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	token := "valid_token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewValidateTokenCommand(
		token,
		xApplication,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, xApplication, correlationID).Return(nil)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_InvalidToken(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewValidateTokenUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	token := "invalid_token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewValidateTokenCommand(
		token,
		xApplication,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, xApplication, correlationID).Return(assert.AnError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewValidateTokenUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	token := "valid_token"
	xApplication := "invalid-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewValidateTokenCommand(
		token,
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
	// AuthClient NÃO deve ser chamado se CompanyClient falhar
	mockAuthClient.AssertNotCalled(t, "ValidateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestValidateTokenUseCase_Execute_EmptyToken(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewValidateTokenUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx := context.Background()
	token := ""
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewValidateTokenCommand(
		token,
		xApplication,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, xApplication, correlationID).Return(nil)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err) // ValidateTokenUseCase não valida token vazio, apenas repassa para o cliente
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	useCase := NewValidateTokenUseCase(mockAuthClient, mockCompanyClient, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela o contexto imediatamente

	token := "valid_token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewValidateTokenCommand(
		token,
		xApplication,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, xApplication, correlationID)
	mockAuthClient.On("ValidateToken", ctx, token, xApplication, correlationID).Return(context.Canceled)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}
