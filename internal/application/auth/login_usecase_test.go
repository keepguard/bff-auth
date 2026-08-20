package auth

import (
	"context"
	"testing"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthClient é um mock para AuthClient
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(inboundDto.AuthResponseDTO), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID string) (inboundDto.RefreshTokenResponseDTO, error) {
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

func (m *MockAuthClient) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID string) error {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID string) error {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(outboundDto.GenerateResetTokenMSResponseDTO), args.Error(1)
}

// MockCompanyClient é um mock para CompanyClient
type MockCompanyClient struct {
	mock.Mock
}

func (m *MockCompanyClient) GetByTenantId(ctx context.Context, tenantId, correlationID string) (authclient.CompanySimpleResponseDTO, error) {
	args := m.Called(ctx, tenantId, correlationID)
	return args.Get(0).(authclient.CompanySimpleResponseDTO), args.Error(1)
}

// setupCompanyMock configura o mock do CompanyClient com resposta de sucesso
func setupCompanyMock(mockCompanyClient *MockCompanyClient, ctx context.Context, tenantId, correlationID string) {
	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		TenantId: tenantId,
		Name:         "Empresa Teste",
		LegalName:    "Empresa Teste LTDA",
		CNPJ:         "12345678000199",
		Status:       "ACTIVE",
	}
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
}

func TestLoginUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"testuser",
		"testpass",
		tenantId,
		correlationID,
		ctx,
	)

	expectedResponse := inboundDto.AuthResponseDTO{
		Token:     "access_token",
		ExpiresIn: 3600,
	}

	expectedReq := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Login", ctx, expectedReq, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockCompanyClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

func TestLoginUseCase_Execute_EmptyUsername(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"",
		"testpass",
		tenantId,
		correlationID,
		ctx,
	)

	// Como a validação é feita no handler, o use case sempre chama o cliente
	// Vamos simular um erro de credenciais inválidas
	expectedReq := inboundDto.AuthRequestDTO{
		Username: "",
		Password: "testpass",
	}

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		TenantId: tenantId,
		Name:         "Empresa Teste",
		LegalName:    "Empresa Teste LTDA",
		CNPJ:         "12345678000199",
		Status:       "ACTIVE",
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockAuthClient.On("Login", ctx, expectedReq, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, assert.AnError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockAuthClient.AssertExpectations(t)
}

func TestLoginUseCase_Execute_EmptyPassword(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"testuser",
		"",
		tenantId,
		correlationID,
		ctx,
	)

	// Como a validação é feita no handler, o use case sempre chama o cliente
	// Vamos simular um erro de credenciais inválidas
	expectedReq := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Login", ctx, expectedReq, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, assert.AnError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockAuthClient.AssertExpectations(t)
}

func TestLoginUseCase_Execute_EmptyUsernameAndPassword(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"",
		"",
		tenantId,
		correlationID,
		ctx,
	)

	// Como a validação é feita no handler, o use case sempre chama o cliente
	// Vamos simular um erro de credenciais inválidas
	expectedReq := inboundDto.AuthRequestDTO{
		Username: "",
		Password: "",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Login", ctx, expectedReq, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, assert.AnError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockAuthClient.AssertExpectations(t)
}

func TestLoginUseCase_Execute_AuthServiceError(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"testuser",
		"testpass",
		tenantId,
		correlationID,
		ctx,
	)

	expectedReq := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Login", ctx, expectedReq, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, assert.AnError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockAuthClient.AssertExpectations(t)
}

func TestLoginUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela o contexto imediatamente

	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"testuser",
		"testpass",
		tenantId,
		correlationID,
		ctx,
	)

	expectedReq := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Login", ctx, expectedReq, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, context.Canceled)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockAuthClient.AssertExpectations(t)
}

func TestLoginUseCase_Execute_InvalidCredentials(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"invaliduser",
		"wrongpass",
		tenantId,
		correlationID,
		ctx,
	)

	expectedReq := inboundDto.AuthRequestDTO{
		Username: "invaliduser",
		Password: "wrongpass",
	}

	authError := &appdto.HTTPError{
		StatusCode: 401,
		Message:    "Credenciais inválidas",
		ErrorCode:  "INVALID_CREDENTIALS",
	}

	// Configurar mocks
	setupCompanyMock(mockCompanyClient, ctx, tenantId, correlationID)
	mockAuthClient.On("Login", ctx, expectedReq, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, authError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &appdto.HTTPError{}, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockAuthClient.AssertExpectations(t)
}

func TestLoginUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := new(MockAuthClient)
	mockCompanyClient := new(MockCompanyClient)
	useCase := NewLoginUseCase(mockAuthClient, mockCompanyClient)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	command := appdto.NewLoginCommand(
		"testuser",
		"testpass",
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
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockCompanyClient.AssertExpectations(t)
	// AuthClient NÃO deve ser chamado se CompanyClient falhar
	mockAuthClient.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
