package message

import (
	"context"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/out/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
	outport "github.com/keepguard/bff-auth/internal/application/port/out"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func strPtr(s string) *string { return &s }

// Reset de conta com idiomas diferentes: o idioma sai normalizado no Locale da
// mensagem, e o valor cru nunca aparece nas variáveis do template.
func TestSendResetPasswordMessageUseCase_LocaleFromAccount(t *testing.T) {
	cases := []struct {
		name string
		raw  *string
		want string
	}{
		{"en-US", strPtr("en-US"), "en-US"},
		{"es-ES", strPtr("es-ES"), "es-ES"},
		{"pt-BR", strPtr("pt-BR"), "pt-BR"},
		{"sem idioma (nil)", nil, ""},
		{"vazio", strPtr(""), ""},
		{"it fica em en-US", strPtr("it"), "en-US"},
		{"ja fica em en-US", strPtr("ja"), "en-US"},
		{"de fica em en-US", strPtr("de"), "en-US"},
		{"pt-PT fica em pt-BR", strPtr("pt-PT"), "pt-BR"},
		{"es-MX fica em es-ES", strPtr("es-MX"), "es-ES"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			email := "test@example.com"
			tenantId := "test-app-uuid"
			correlationID := "corr-1"

			mockUserClient := new(MockUserClient)
			mockCompanyClient := new(MockCompanyClient)
			mockMessagePublisher := new(MockMessagePublisher)
			mockAuthClient := new(MockAuthClient)
			useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, zap.NewNop())

			mockCompanyClient.On("GetByTenantId", ctx, tenantId, correlationID).Return(authclient.CompanySimpleResponseDTO{ID: "company-123"}, nil)
			mockUserClient.On("GetByEmail", ctx, email, tenantId, "company-123", correlationID).Return(outboundDto.UserByEmailResponseDTO{
				ID: "user-123", CodeUser: "code-user-123", Username: "testuser", Email: email, Status: "ACTIVE", PreferredLocale: tc.raw,
			}, nil)
			mockAuthClient.On("GenerateResetToken", ctx, mock.Anything, tenantId, correlationID).Return(outboundDto.GenerateResetTokenMSResponseDTO{Token: "123456"}, nil)

			var published outport.MessageDTO
			mockMessagePublisher.On("PublishMessage", ctx, mock.Anything).Run(func(args mock.Arguments) {
				published = args.Get(1).(outport.MessageDTO)
			}).Return(nil)

			res, err := useCase.Execute(ctx, appdto.NewSendResetPasswordMessageCommand(email, tenantId, correlationID))

			assert.NoError(t, err)
			assert.True(t, res.Success)
			assert.Equal(t, tc.want, published.Locale)
			assert.Equal(t, map[string]interface{}{"userName": "testuser", "token": "123456"}, published.Variables)
		})
	}
}

// Falha de publicação não muda a resposta: o envio segue como antes, com ou sem idioma.
func TestSendResetPasswordMessageUseCase_LocaleDoesNotChangeResponse(t *testing.T) {
	ctx := context.Background()
	mockUserClient := new(MockUserClient)
	mockCompanyClient := new(MockCompanyClient)
	mockMessagePublisher := new(MockMessagePublisher)
	mockAuthClient := new(MockAuthClient)
	useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, zap.NewNop())

	mockCompanyClient.On("GetByTenantId", ctx, "t", "c").Return(authclient.CompanySimpleResponseDTO{ID: "company-123"}, nil)
	mockUserClient.On("GetByEmail", ctx, "a@b.com", "t", "company-123", "c").Return(outboundDto.UserByEmailResponseDTO{
		CodeUser: "u", Username: "x", Status: "ACTIVE", PreferredLocale: strPtr("en-US"),
	}, nil)
	mockAuthClient.On("GenerateResetToken", ctx, mock.Anything, "t", "c").Return(outboundDto.GenerateResetTokenMSResponseDTO{Token: "1"}, nil)
	mockMessagePublisher.On("PublishMessage", ctx, mock.Anything).Return(assertErr{})

	res, err := useCase.Execute(ctx, appdto.NewSendResetPasswordMessageCommand("a@b.com", "t", "c"))

	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, appdto.ResetPasswordGenericMessage, res.Message)
}

type assertErr struct{}

func (assertErr) Error() string { return "broker indisponível" }

// Sem idioma guardado na conta, vale o idioma que a pessoa usa na tela (enviado pelo cliente); com idioma
// guardado, a conta vence. Valor cru nunca passa sem normalizar.
func TestSendResetPasswordMessageUseCase_LocaleFromRequest(t *testing.T) {
	cases := []struct {
		name    string
		account *string
		request string
		want    string
	}{
		{"conta sem idioma usa o da tela", nil, "en-US", "en-US"},
		{"conta com idioma vazio usa o da tela", strPtr(""), "es-MX", "es-ES"},
		{"conta vence a tela", strPtr("pt-BR"), "en-US", "pt-BR"},
		{"tela fora do conjunto vira en-US", nil, "ja", "en-US"},
		{"sem nenhum idioma sai vazio", nil, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			email, tenantId, corr := "test@example.com", "test-app-uuid", "corr-1"
			mockUserClient := new(MockUserClient)
			mockCompanyClient := new(MockCompanyClient)
			mockMessagePublisher := new(MockMessagePublisher)
			mockAuthClient := new(MockAuthClient)
			useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockMessagePublisher, zap.NewNop())
			mockCompanyClient.On("GetByTenantId", ctx, tenantId, corr).Return(authclient.CompanySimpleResponseDTO{ID: "company-123"}, nil)
			mockUserClient.On("GetByEmail", ctx, email, tenantId, "company-123", corr).Return(outboundDto.UserByEmailResponseDTO{
				ID: "user-123", CodeUser: "code-user-123", Username: "testuser", Email: email, Status: "ACTIVE", PreferredLocale: tc.account,
			}, nil)
			mockAuthClient.On("GenerateResetToken", ctx, mock.Anything, tenantId, corr).Return(outboundDto.GenerateResetTokenMSResponseDTO{Token: "123456"}, nil)
			var published outport.MessageDTO
			mockMessagePublisher.On("PublishMessage", ctx, mock.Anything).Run(func(args mock.Arguments) {
				published = args.Get(1).(outport.MessageDTO)
			}).Return(nil)

			cmd := appdto.NewSendResetPasswordMessageCommand(email, tenantId, corr)
			cmd.Locale = tc.request
			res, err := useCase.Execute(ctx, cmd)
			assert.NoError(t, err)
			assert.True(t, res.Success)
			assert.Equal(t, tc.want, published.Locale)
		})
	}
}
