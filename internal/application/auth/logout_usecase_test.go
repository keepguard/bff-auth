package auth

import (
	"context"
	"testing"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogoutUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLogoutUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	token := "valid_access_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLogoutCommand(
		token,
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Logout", ctx, token, tenantId, correlationID).Return(nil)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_AuthServiceError(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLogoutUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	token := "invalid_access_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLogoutCommand(
		token,
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Logout", ctx, token, tenantId, correlationID).Return(assert.AnError)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLogoutUseCase(mockAuthClient, mockCompanyClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela o contexto imediatamente

	token := "valid_access_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLogoutCommand(
		token,
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Logout", ctx, token, tenantId, correlationID).Return(context.Canceled)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_EmptyToken(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLogoutUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	token := ""
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLogoutCommand(
		token,
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Logout", ctx, token, tenantId, correlationID).Return(nil)

	// Act
	err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err) // LogoutUseCase não valida token vazio, apenas repassa para o cliente
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLogoutUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	token := "valid_access_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLogoutCommand(
		token,
		tenantId,
		correlationID,
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
	mockAuthClient.AssertNotCalled(t, "Logout", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
