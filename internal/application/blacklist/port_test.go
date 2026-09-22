package blacklist

import (
	"context"
	"errors"
	"net/http"
	"testing"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/port"
	"github.com/stretchr/testify/mock"
)

type mockAuthClient struct {
	mock.Mock
	port.AuthClient
}

func (m *mockAuthClient) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]appdto.DeviceBlacklistDTO, error) {
	args := m.Called(ctx, token, tenantId, correlationID)
	return args.Get(0).([]appdto.DeviceBlacklistDTO), args.Error(1)
}

func (m *mockAuthClient) AddDeviceToBlacklist(ctx context.Context, req appdto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return m.Called(ctx, req, token, tenantId, correlationID).Error(0)
}

func (m *mockAuthClient) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	return m.Called(ctx, deviceId, token, tenantId, correlationID).Error(0)
}

func (m *mockAuthClient) SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (appdto.PaginatedDeviceBlacklistResponseDTO, error) {
	args := m.Called(ctx, queryParams, token, tenantId, correlationID)
	return args.Get(0).(appdto.PaginatedDeviceBlacklistResponseDTO), args.Error(1)
}

func (m *mockAuthClient) AdminAddDeviceToBlacklist(ctx context.Context, req appdto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return m.Called(ctx, req, token, tenantId, correlationID).Error(0)
}

func (m *mockAuthClient) AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error {
	return m.Called(ctx, deviceId, userId, token, tenantId, correlationID).Error(0)
}

func TestListMe(t *testing.T) {
	cli := &mockAuthClient{}
	esperado := []appdto.DeviceBlacklistDTO{{DeviceID: "d-1"}}
	cli.On("ListDeviceBlacklist", mock.Anything, "tok", "t1", "c1").Return(esperado, nil)

	svc := NewBlacklistPort(cli)
	got, err := svc.ListMe(context.Background(), appdto.ListDeviceBlacklistQuery{
		Token: "tok", TenantID: "t1", CorrelationID: "c1",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got) != 1 || got[0].DeviceID != "d-1" {
		t.Errorf("resultado = %+v", got)
	}
	cli.AssertExpectations(t)
}

func TestListMe_ErroDoClientePropaga(t *testing.T) {
	cli := &mockAuthClient{}
	cli.On("ListDeviceBlacklist", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]appdto.DeviceBlacklistDTO{}, errors.New("ms fora"))

	svc := NewBlacklistPort(cli)
	if _, err := svc.ListMe(context.Background(), appdto.ListDeviceBlacklistQuery{}); err == nil {
		t.Fatal("erro do cliente foi engolido")
	}
}

// Bloquear um dispositivo é ação de segurança: o comando tem que chegar ao
// ms com os campos que o usuário preencheu, não com um request vazio.
func TestAddMe_MontaORequestComOsCamposDoComando(t *testing.T) {
	cli := &mockAuthClient{}
	cli.On("AddDeviceToBlacklist", mock.Anything,
		appdto.AddDeviceBlacklistRequestDTO{
			DeviceID: "d-1", DeviceName: "Celular", Reason: "perdido",
		}, "tok", "t1", "c1").Return(nil)

	svc := NewBlacklistPort(cli)
	err := svc.AddMe(context.Background(), appdto.AddDeviceBlacklistCommand{
		DeviceID: "d-1", DeviceName: "Celular", Reason: "perdido",
		Token: "tok", TenantID: "t1", CorrelationID: "c1",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	cli.AssertExpectations(t)
}

func TestRemoveMe(t *testing.T) {
	cli := &mockAuthClient{}
	cli.On("RemoveDeviceFromBlacklist", mock.Anything, "d-1", "tok", "t1", "c1").Return(nil)

	svc := NewBlacklistPort(cli)
	err := svc.RemoveMe(context.Background(), appdto.RemoveDeviceBlacklistCommand{
		DeviceID: "d-1", Token: "tok", TenantID: "t1", CorrelationID: "c1",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	cli.AssertExpectations(t)
}

func TestAdminAdd(t *testing.T) {
	cli := &mockAuthClient{}
	cli.On("AdminAddDeviceToBlacklist", mock.Anything, mock.Anything, "tok", "t1", "c1").Return(nil)

	svc := NewBlacklistPort(cli)
	err := svc.AdminAdd(context.Background(), appdto.AdminAddDeviceBlacklistCommand{
		DeviceID: "d-1", Token: "tok", TenantID: "t1", CorrelationID: "c1",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	cli.AssertExpectations(t)
}

func TestAdminRemove(t *testing.T) {
	cli := &mockAuthClient{}
	cli.On("AdminRemoveDeviceFromBlacklist", mock.Anything, "d-1", "u-1", "tok", "t1", "c1").Return(nil)

	svc := NewBlacklistPort(cli)
	err := svc.AdminRemove(context.Background(), appdto.AdminRemoveDeviceBlacklistCommand{
		DeviceID: "d-1", UserID: "u-1", Token: "tok", TenantID: "t1", CorrelationID: "c1",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	cli.AssertExpectations(t)
}

func TestSearchAdmin(t *testing.T) {
	cli := &mockAuthClient{}
	esperado := appdto.PaginatedDeviceBlacklistResponseDTO{}
	cli.On("SearchAdminDeviceBlacklist", mock.Anything, mock.Anything, "tok", "t1", "c1").
		Return(esperado, nil)

	svc := NewBlacklistPort(cli)
	_, err := svc.SearchAdmin(context.Background(), appdto.SearchAdminDeviceBlacklistQuery{
		Token: "tok", TenantID: "t1", CorrelationID: "c1",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	cli.AssertExpectations(t)
}

// Sem cliente configurado, cada operação devolve 503 — e não panic. É o que
// mantém o BFF respondendo quando o ms-auth não subiu.
func TestSemClienteDevolve503(t *testing.T) {
	svc := NewBlacklistPort(nil)
	ctx := context.Background()

	casos := map[string]func() error{
		"ListMe": func() error {
			_, err := svc.ListMe(ctx, appdto.ListDeviceBlacklistQuery{})
			return err
		},
		"AddMe": func() error {
			return svc.AddMe(ctx, appdto.AddDeviceBlacklistCommand{})
		},
		"RemoveMe": func() error {
			return svc.RemoveMe(ctx, appdto.RemoveDeviceBlacklistCommand{})
		},
		"SearchAdmin": func() error {
			_, err := svc.SearchAdmin(ctx, appdto.SearchAdminDeviceBlacklistQuery{})
			return err
		},
		"AdminAdd": func() error {
			return svc.AdminAdd(ctx, appdto.AdminAddDeviceBlacklistCommand{})
		},
		"AdminRemove": func() error {
			return svc.AdminRemove(ctx, appdto.AdminRemoveDeviceBlacklistCommand{})
		},
	}

	for nome, chama := range casos {
		t.Run(nome, func(t *testing.T) {
			err := chama()
			if err == nil {
				t.Fatal("não sinalizou indisponibilidade")
			}
			var appErr interface{ StatusCode() int }
			if errors.As(err, &appErr) {
				if appErr.StatusCode() != http.StatusServiceUnavailable {
					t.Errorf("status = %d, quer 503", appErr.StatusCode())
				}
			}
		})
	}
}
