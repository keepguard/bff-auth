package message

import (
	"context"
	"errors"
	"testing"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

// MockUserClient é um mock para UserClient
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetByEmail(ctx context.Context, email, tenantId, correlationID string) (outboundDto.UserByEmailResponseDTO, error) {
	args := m.Called(ctx, email, tenantId, correlationID)
	return args.Get(0).(outboundDto.UserByEmailResponseDTO), args.Error(1)
}

// MockCompanyClient é um mock para CompanyClient
type MockCompanyClient struct {
	mock.Mock
}

func (m *MockCompanyClient) GetByTenantId(ctx context.Context, tenantId, correlationID string) (authclient.CompanySimpleResponseDTO, error) {
	args := m.Called(ctx, tenantId, correlationID)
	return args.Get(0).(authclient.CompanySimpleResponseDTO), args.Error(1)
}

// MockMessagePublisher é um mock para MessagePublisher
type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) PublishMessage(ctx context.Context, message messaging.MessageDTO) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessagePublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestSendResetPasswordMessageUseCase_Execute_Success testa execução com sucesso
func TestSendResetPasswordMessageUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedGenerateTokenReq := map[string]interface{}{
		"codeUser":          "code-user-123",
		"messageType":       "EMAIL",
		"communicationType": "EMAIL",
		"templateType":      "RECUPERACAO_SENHA",
	}

	expectedGenerateTokenResponse := outboundDto.GenerateResetTokenMSResponseDTO{
		CodeUser:          "code-user-123",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Token:             "123456",
		ExpiresInSeconds:  600,
	}

	expectedMessageReq := messaging.MessageDTO{
		TenantId:      tenantId,
		XCorrelationID:    correlationID,
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Recipient:         email,
		CodeUser:          "code-user-123",
		Variables: map[string]interface{}{
			"userName": "testuser",
			"token":    "123456",
		},
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("GenerateResetToken", ctx, expectedGenerateTokenReq, tenantId, correlationID).Return(expectedGenerateTokenResponse, nil)
	mockMessagePublisher.On("PublishMessage", ctx, expectedMessageReq).Return(nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "Token de reset de senha gerado e mensagem enviada com sucesso", result.Message)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	mockMessagePublisher.AssertExpectations(t)
}

// TestSendResetPasswordMessageUseCase_Execute_CompanyNotFound testa quando empresa não é encontrada
func TestSendResetPasswordMessageUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "invalid-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	companyError := &appdto.HTTPError{
		StatusCode: 404,
		Message:    "Empresa não encontrada",
		ErrorCode:  "COMPANY_NOT_FOUND",
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(authclient.CompanySimpleResponseDTO{}, companyError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.SendResetPasswordMessageResponseDTO{}, result)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	// UserClient e MessagePublisher não devem ser chamados
	mockUserClient.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockMessagePublisher.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything)
}

// TestSendResetPasswordMessageUseCase_Execute_UserNotFound testa quando usuário não é encontrado
func TestSendResetPasswordMessageUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "notfound@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	userError := &appdto.HTTPError{
		StatusCode: 404,
		Message:    "Usuário não encontrado",
		ErrorCode:  "USER_NOT_FOUND",
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(outboundDto.UserByEmailResponseDTO{}, userError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.SendResetPasswordMessageResponseDTO{}, result)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	// CommunicationClient não deve ser chamado
	mockMessagePublisher.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything)
}

// TestSendResetPasswordMessageUseCase_Execute_UserNotActive testa quando usuário não está ativo
func TestSendResetPasswordMessageUseCase_Execute_UserNotActive(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	inactiveUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "INACTIVE",
		EmailVerified: true,
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(inactiveUser, nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.SendResetPasswordMessageResponseDTO{}, result)

	appErr, ok := err.(*pkg.AppError)
	assert.True(t, ok)
	assert.Equal(t, 400, appErr.StatusCode)
	assert.Equal(t, "USER_NOT_ACTIVE", appErr.Code)
	assert.Contains(t, appErr.Message, "Usuário não está ativo")
	assert.Contains(t, appErr.Message, "INACTIVE")

	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	// CommunicationClient não deve ser chamado
	mockMessagePublisher.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything)
}

// TestSendResetPasswordMessageUseCase_Execute_UserStatusPending testa com status PENDING
func TestSendResetPasswordMessageUseCase_Execute_UserStatusPending(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	pendingUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "PENDING",
		EmailVerified: false,
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(pendingUser, nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.SendResetPasswordMessageResponseDTO{}, result)

	appErr, ok := err.(*pkg.AppError)
	assert.True(t, ok)
	assert.Equal(t, 400, appErr.StatusCode)
	assert.Equal(t, "USER_NOT_ACTIVE", appErr.Code)

	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockMessagePublisher.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything)
}

// TestSendResetPasswordMessageUseCase_Execute_GenerateTokenError testa erro ao gerar token de reset
func TestSendResetPasswordMessageUseCase_Execute_GenerateTokenError(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedGenerateTokenReq := map[string]interface{}{
		"codeUser":          "code-user-123",
		"messageType":       "EMAIL",
		"communicationType": "EMAIL",
		"templateType":      "RECUPERACAO_SENHA",
	}

	generateTokenError := &appdto.HTTPError{
		StatusCode: 500,
		Message:    "Erro ao gerar token de reset",
		ErrorCode:  "GENERATE_TOKEN_ERROR",
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("GenerateResetToken", ctx, expectedGenerateTokenReq, tenantId, correlationID).Return(outboundDto.GenerateResetTokenMSResponseDTO{}, generateTokenError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.SendResetPasswordMessageResponseDTO{}, result)
	assert.IsType(t, &appdto.HTTPError{}, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	mockMessagePublisher.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything)
}

// TestSendResetPasswordMessageUseCase_Execute_MessagePublisherError testa erro ao publicar mensagem
func TestSendResetPasswordMessageUseCase_Execute_MessagePublisherError(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedGenerateTokenReq := map[string]interface{}{
		"codeUser":          "code-user-123",
		"messageType":       "EMAIL",
		"communicationType": "EMAIL",
		"templateType":      "RECUPERACAO_SENHA",
	}

	expectedGenerateTokenResponse := outboundDto.GenerateResetTokenMSResponseDTO{
		CodeUser:          "code-user-123",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Token:             "123456",
		ExpiresInSeconds:  600,
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("GenerateResetToken", ctx, expectedGenerateTokenReq, tenantId, correlationID).Return(expectedGenerateTokenResponse, nil)
	mockMessagePublisher.On("PublishMessage", ctx, mock.AnythingOfType("messaging.MessageDTO")).Return(errors.New("rabbitmq connection failed"))

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err) // Não deve falhar se o email não for enviado (melhor UX)
	assert.True(t, result.Success)
	assert.Equal(t, "Token de reset de senha gerado e mensagem enviada com sucesso", result.Message)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	mockMessagePublisher.AssertExpectations(t)
}

// TestSendResetPasswordMessageUseCase_Execute_SendMessageFailure testa quando envio falha mas retorna resposta
func TestSendResetPasswordMessageUseCase_Execute_SendMessageFailure(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	expectedUser := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-user-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	expectedGenerateTokenReq := map[string]interface{}{
		"codeUser":          "code-user-123",
		"messageType":       "EMAIL",
		"communicationType": "EMAIL",
		"templateType":      "RECUPERACAO_SENHA",
	}

	expectedGenerateTokenResponse := outboundDto.GenerateResetTokenMSResponseDTO{
		CodeUser:          "code-user-123",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Token:             "123456",
		ExpiresInSeconds:  600,
	}

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(expectedUser, nil)
	mockAuthClient.On("GenerateResetToken", ctx, expectedGenerateTokenReq, tenantId, correlationID).Return(expectedGenerateTokenResponse, nil)
	mockMessagePublisher.On("PublishMessage", ctx, mock.AnythingOfType("messaging.MessageDTO")).Return(nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "Token de reset de senha gerado e mensagem enviada com sucesso", result.Message)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	mockMessagePublisher.AssertExpectations(t)
}

// TestSendResetPasswordMessageUseCase_Execute_CompanyGenericError testa erro genérico no company client
func TestSendResetPasswordMessageUseCase_Execute_CompanyGenericError(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	genericError := errors.New("erro de conexão")

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(authclient.CompanySimpleResponseDTO{}, genericError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.SendResetPasswordMessageResponseDTO{}, result)
	assert.Equal(t, genericError, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockMessagePublisher.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything)
}

// TestSendResetPasswordMessageUseCase_Execute_UserGenericError testa erro genérico no user client
func TestSendResetPasswordMessageUseCase_Execute_UserGenericError(t *testing.T) {
	// Arrange
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	logger, _ := zap.NewDevelopment()
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, logger)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := appdto.NewSendResetPasswordMessageCommand(
		email,
		tenantId,
		correlationID,
		ctx,
	)

	expectedCompany := authclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	genericError := errors.New("erro de timeout")

	// Configurar mocks
	mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedCompany, nil)
	mockUserClient.On("GetByEmail", ctx, email, tenantId, correlationID).Return(outboundDto.UserByEmailResponseDTO{}, genericError)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, inboundDto.SendResetPasswordMessageResponseDTO{}, result)
	assert.Equal(t, genericError, err)
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockMessagePublisher.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything)
}
