package message

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/out/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Tabela por status no envio da mensagem de recuperação: só ACTIVE gera o token; os demais recebem USER_NOT_ACTIVE.
func TestSendResetPasswordMessage_StatusTable(t *testing.T) {
	cases := []struct {
		status    string
		reachesMS bool
	}{
		{status: "ACTIVE", reachesMS: true},
		{status: "PENDING_ACTIVATION", reachesMS: false},
		{status: "PENDING", reachesMS: false},
		{status: "BLOCKED", reachesMS: false},
		{status: "DELETED", reachesMS: false},
	}
	for _, tc := range cases {
		t.Run("status="+tc.status, func(t *testing.T) {
			ctx := context.Background()
			const tenantId, corr, email = "test-app-uuid", "corr-1", "test@example.com"
			mockUserClient := new(MockUserClient)
			mockCompanyClient := new(MockCompanyClient)
			mockAuthClient := new(MockAuthClient)
			useCase := NewSendResetPasswordMessageUseCase(mockAuthClient, mockUserClient, mockCompanyClient, new(MockMessagePublisher), zap.NewNop())

			mockCompanyClient.On("GetByTenantId", ctx, tenantId, corr).Return(authclient.CompanySimpleResponseDTO{ID: "company-123", Name: "Test"}, nil)
			mockUserClient.On("GetByEmail", ctx, email, tenantId, "company-123", corr).Return(outboundDto.UserByEmailResponseDTO{
				ID: "user-123", CodeUser: "code-user-123", Email: email, Status: tc.status,
			}, nil)
			if tc.reachesMS {
				mockAuthClient.On("GenerateResetToken", mock.Anything, mock.Anything, tenantId, corr).Return(outboundDto.GenerateResetTokenMSResponseDTO{}, errors.New("parada no ms-auth"))
			}

			_, err := useCase.Execute(ctx, appdto.NewSendResetPasswordMessageCommand(email, tenantId, corr))

			if tc.reachesMS {
				assert.NotEqual(t, "USER_NOT_ACTIVE", errCodeOf(err))
				mockAuthClient.AssertCalled(t, "GenerateResetToken", mock.Anything, mock.Anything, tenantId, corr)
				return
			}
			appErr, ok := err.(*pkg.AppError)
			if assert.True(t, ok, "esperava AppError de status") {
				assert.Equal(t, "USER_NOT_ACTIVE", appErr.Code)
				assert.Equal(t, 400, appErr.StatusCode)
			}
			mockAuthClient.AssertNotCalled(t, "GenerateResetToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func errCodeOf(err error) string {
	if appErr, ok := err.(*pkg.AppError); ok {
		return appErr.Code
	}
	return ""
}
