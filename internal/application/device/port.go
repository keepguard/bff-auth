package device

import (
	"context"
	"net/http"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/pkg"
)

type DevicePort interface {
	SendChallenge(ctx context.Context, cmd appdto.SendDeviceChallengeCommand) (map[string]any, error)
	VerifyChallenge(ctx context.Context, cmd appdto.VerifyDeviceChallengeCommand) (appdto.AuthResponseDTO, error)
	QuickRevoke(ctx context.Context, cmd appdto.QuickRevokeCommand) (appdto.QuickRevokeViewDTO, error)
}

type service struct {
	auth port.AuthClient
}

func NewDevicePort(authClient port.AuthClient) DevicePort {
	return &service{auth: authClient}
}

func (s *service) SendChallenge(ctx context.Context, cmd appdto.SendDeviceChallengeCommand) (map[string]any, error) {
	if s == nil || s.auth == nil {
		return nil, pkg.NewAppError("SERVICE_UNAVAILABLE", "Desafio de dispositivo indisponível", http.StatusServiceUnavailable)
	}
	return s.auth.SendDeviceChallenge(ctx, appdto.DeviceChallengeSendRequestDTO{
		ChallengeSessionID: cmd.ChallengeSessionID,
		Channel:            cmd.Channel,
	}, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) VerifyChallenge(ctx context.Context, cmd appdto.VerifyDeviceChallengeCommand) (appdto.AuthResponseDTO, error) {
	if s == nil || s.auth == nil {
		return appdto.AuthResponseDTO{}, pkg.NewAppError("SERVICE_UNAVAILABLE", "Desafio de dispositivo indisponível", http.StatusServiceUnavailable)
	}
	return s.auth.VerifyDeviceChallenge(ctx, appdto.DeviceChallengeVerifyRequestDTO{
		ChallengeSessionID: cmd.ChallengeSessionID,
		Code:               cmd.Code,
		TrustDevice:        cmd.TrustDevice,
	}, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) QuickRevoke(ctx context.Context, cmd appdto.QuickRevokeCommand) (appdto.QuickRevokeViewDTO, error) {
	if s == nil || s.auth == nil {
		return appdto.QuickRevokeViewDTO{}, pkg.NewAppError("SERVICE_UNAVAILABLE", "Revogação rápida indisponível", http.StatusServiceUnavailable)
	}
	payload, err := s.auth.QuickRevoke(ctx, cmd.Token, cmd.Blacklist, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.QuickRevokeViewDTO{}, err
	}
	return appdto.QuickRevokeFromPayload(payload), nil
}
