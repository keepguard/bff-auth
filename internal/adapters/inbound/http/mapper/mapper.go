package mapper

import (
	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
)

func ToSendDeviceChallengeCommand(req inboundDto.DeviceChallengeSendRequestDTO, tenantID, correlationID string) appdto.SendDeviceChallengeCommand {
	return appdto.SendDeviceChallengeCommand{
		TenantID:           tenantID,
		CorrelationID:      correlationID,
		ChallengeSessionID: req.ChallengeSessionID,
		Channel:            req.Channel,
	}
}

func ToVerifyDeviceChallengeCommand(req inboundDto.DeviceChallengeVerifyRequestDTO, tenantID, correlationID string) appdto.VerifyDeviceChallengeCommand {
	return appdto.VerifyDeviceChallengeCommand{
		TenantID:           tenantID,
		CorrelationID:      correlationID,
		ChallengeSessionID: req.ChallengeSessionID,
		Code:               req.Code,
		TrustDevice:        req.TrustDevice,
	}
}

func ToQuickRevokeCommand(token, tenantID, correlationID string, blacklist bool) appdto.QuickRevokeCommand {
	return appdto.QuickRevokeCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Token:         token,
		Blacklist:     blacklist,
	}
}

func ToAuthResponse(view appdto.AuthResponseDTO) inboundDto.AuthResponseDTO {
	return view
}

func ToRefreshTokenResponse(view appdto.RefreshTokenResponseDTO) inboundDto.RefreshTokenResponseDTO {
	return view
}

func ToSendResetPasswordMessageResponse(view appdto.SendResetPasswordMessageViewDTO) inboundDto.SendResetPasswordMessageResponseDTO {
	return inboundDto.SendResetPasswordMessageResponseDTO{
		Success: view.Success,
		Message: view.Message,
	}
}

func ToAddDeviceBlacklistCommand(req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantID, correlationID string) appdto.AddDeviceBlacklistCommand {
	return appdto.AddDeviceBlacklistCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Token:         token,
		DeviceID:      req.DeviceID,
		DeviceName:    req.DeviceName,
		Reason:        req.Reason,
	}
}

func ToAdminAddDeviceBlacklistCommand(req inboundDto.AdminAddDeviceBlacklistRequestDTO, token, tenantID, correlationID string) appdto.AdminAddDeviceBlacklistCommand {
	return appdto.AdminAddDeviceBlacklistCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Token:         token,
		UserID:        req.UserID,
		DeviceID:      req.DeviceID,
		DeviceName:    req.DeviceName,
		Reason:        req.Reason,
		ExpiresAt:     req.ExpiresAt,
	}
}

func ToTenantAddDeviceBlacklistCommand(req inboundDto.AddDeviceBlacklistRequestDTO, userID, token, tenantID, correlationID string) appdto.AdminAddDeviceBlacklistCommand {
	return appdto.AdminAddDeviceBlacklistCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Token:         token,
		UserID:        userID,
		DeviceID:      req.DeviceID,
		DeviceName:    req.DeviceName,
		Reason:        req.Reason,
	}
}

func QuickRevokeJSON(view appdto.QuickRevokeViewDTO) map[string]any {
	if view.Payload != nil {
		return view.Payload
	}
	payload := map[string]any{}
	if view.Message != "" {
		payload["message"] = view.Message
	}
	return payload
}
