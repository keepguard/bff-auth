package lifecycle

import (
	"context"
	"net/http"
	"testing"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthClient struct {
	mock.Mock
	port.AuthClient
}

func (m *mockAuthClient) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (appdto.UserByCodeResponseDTO, error) {
	args := m.Called(ctx, codeUser, token, tenantId, correlationID)
	return args.Get(0).(appdto.UserByCodeResponseDTO), args.Error(1)
}

func (m *mockAuthClient) BlockUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	return m.Called(ctx, idUserExternal, reason, token, tenantId, correlationID).Error(0)
}

func (m *mockAuthClient) DeleteUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	return m.Called(ctx, idUserExternal, reason, token, tenantId, correlationID).Error(0)
}

type mockCompanyClient struct {
	mock.Mock
}

func (m *mockCompanyClient) GetByTenantId(ctx context.Context, tenantId, correlationID string) (appdto.CompanySimpleResponseDTO, error) {
	args := m.Called(ctx, tenantId, correlationID)
	return args.Get(0).(appdto.CompanySimpleResponseDTO), args.Error(1)
}

func TestLifecyclePort_BlockMe(t *testing.T) {
	authClient := new(mockAuthClient)
	companyClient := new(mockCompanyClient)
	svc := NewLifecyclePort(authClient, companyClient)

	cmd := appdto.AccountLifecycleCommand{
		TenantID:      "tenant-1",
		Token:         "tok",
		CodeUser:      "user-sub-1",
		Reason:        "quero bloquear",
		CorrelationID: "corr-1",
	}

	companyClient.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(appdto.CompanySimpleResponseDTO{ID: "company-1"}, nil)
	authClient.On("GetUserByCodeUser", mock.Anything, "user-sub-1", "tok", "tenant-1", "corr-1").
		Return(appdto.UserByCodeResponseDTO{IDUserExternal: "ext-id-1"}, nil)
	authClient.On("BlockUser", mock.Anything, "ext-id-1", "quero bloquear", "tok", "tenant-1", "corr-1").
		Return(nil)

	err := svc.BlockMe(context.Background(), cmd)
	assert.NoError(t, err)
	authClient.AssertExpectations(t)
	companyClient.AssertExpectations(t)
}

func TestLifecyclePort_DeleteMe_PropagatesError(t *testing.T) {
	authClient := new(mockAuthClient)
	companyClient := new(mockCompanyClient)
	svc := NewLifecyclePort(authClient, companyClient)

	cmd := appdto.AccountLifecycleCommand{
		TenantID:      "tenant-1",
		Token:         "tok",
		CodeUser:      "user-sub-1",
		Reason:        "encerrar",
		CorrelationID: "corr-1",
	}

	companyClient.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(appdto.CompanySimpleResponseDTO{ID: "company-1"}, nil)
	authClient.On("GetUserByCodeUser", mock.Anything, "user-sub-1", "tok", "tenant-1", "corr-1").
		Return(appdto.UserByCodeResponseDTO{IDUserExternal: "ext-id-1"}, nil)
	authClient.On("DeleteUser", mock.Anything, "ext-id-1", "encerrar", "tok", "tenant-1", "corr-1").
		Return(&appdto.HTTPError{StatusCode: http.StatusForbidden, ErrorCode: "FORBIDDEN", Message: "Sem permissão"})

	err := svc.DeleteMe(context.Background(), cmd)
	assert.Error(t, err)
	httpErr, ok := err.(*appdto.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.StatusCode)
}
