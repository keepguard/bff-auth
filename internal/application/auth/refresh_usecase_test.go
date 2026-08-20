package auth

import (
	"context"
	"testing"

	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefreshUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewRefreshUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewRefreshTokenCommand(
		"valid_refresh_token",
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)

	expectedResponse := dto.RefreshTokenResponseDTO{
		Token:     "new_access_token",
		ExpiresIn: 3600,
	}

	expectedReq := dto.RefreshTokenRequestDTO{
		Token: "valid_refresh_token",
	}

	mockAuthClient.On("RefreshToken", ctx, expectedReq, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockCompanyClient.AssertExpectations(t)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_EmptyRefreshToken(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewRefreshUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewRefreshTokenCommand(
		"",
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)

	// Como a validação é feita no handler, o use case sempre chama o cliente
	// Vamos simular um erro
	expectedReq := dto.RefreshTokenRequestDTO{
		Token: "",
	}

	mockAuthClient.On("RefreshToken", ctx, expectedReq, tenantId, correlationID).Return(dto.RefreshTokenResponseDTO{}, assert.AnError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dto.RefreshTokenResponseDTO{}, result)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_AuthServiceError(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewRefreshUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewRefreshTokenCommand(
		"invalid_refresh_token",
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)

	expectedReq := dto.RefreshTokenRequestDTO{
		Token: "invalid_refresh_token",
	}

	mockAuthClient.On("RefreshToken", ctx, expectedReq, tenantId, correlationID).Return(dto.RefreshTokenResponseDTO{}, assert.AnError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dto.RefreshTokenResponseDTO{}, result)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewRefreshUseCase(mockAuthClient, mockCompanyClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela o contexto imediatamente

	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewRefreshTokenCommand(
		"valid_refresh_token",
		tenantId,
		correlationID,
		ctx,
	)

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)

	expectedReq := dto.RefreshTokenRequestDTO{
		Token: "valid_refresh_token",
	}

	mockAuthClient.On("RefreshToken", ctx, expectedReq, tenantId, correlationID).Return(dto.RefreshTokenResponseDTO{}, context.Canceled)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, dto.RefreshTokenResponseDTO{}, result)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewRefreshUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewRefreshTokenCommand(
		"valid_refresh_token",
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
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	assert.Equal(t, dto.RefreshTokenResponseDTO{}, result)
	mockCompanyClient.AssertExpectations(t)
	// AuthClient NÃO deve ser chamado se CompanyClient falhar
	mockAuthClient.AssertNotCalled(t, "RefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
