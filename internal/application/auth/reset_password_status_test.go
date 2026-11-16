package auth

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/out/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Tabela por status no reset de senha: só ACTIVE chega ao ms-auth; os demais recebem USER_NOT_ACTIVE.
func TestResetPassword_StatusTable(t *testing.T) {
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
			mockAuthClient := new(MockAuthClient)
			mockUserClient := new(MockUserClient)
			mockCompanyClient := new(MockCompanyClient)
			useCase := NewResetPasswordUseCase(mockAuthClient, mockUserClient, mockCompanyClient, nil, zap.NewNop())

			setupCompanyMock(mockCompanyClient, ctx, tenantId, corr)
			mockUserClient.On("GetByEmail", ctx, email, tenantId, "550e8400-e29b-41d4-a716-446655440000", corr).Return(outboundDto.UserByEmailResponseDTO{
				ID: "user-123", CodeUser: "code-user-123", Email: email, Status: tc.status,
			}, nil)
			if tc.reachesMS {
				// Stop no ms-auth: o que importa é que a recusa por status não aconteceu.
				mockAuthClient.On("ResetPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("parada no ms-auth"))
			}

			command := appdto.NewResetPasswordCommand(email, "token", "NovaSenha@123", "NovaSenha@123", tenantId, corr, "", "", "", "", "")
			err := useCase.Execute(ctx, command)

			if tc.reachesMS {
				assert.NotEqual(t, "USER_NOT_ACTIVE", errCode(err))
				mockAuthClient.AssertCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				return
			}
			appErr, ok := err.(*pkg.AppError)
			if assert.True(t, ok, "esperava AppError de status") {
				assert.Equal(t, "USER_NOT_ACTIVE", appErr.Code)
				assert.Equal(t, 400, appErr.StatusCode)
			}
			mockAuthClient.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func errCode(err error) string {
	if appErr, ok := err.(*pkg.AppError); ok {
		return appErr.Code
	}
	return ""
}
