package dto

type SendDeviceChallengeCommand struct {
	TenantID           string
	CorrelationID      string
	ChallengeSessionID string
	Channel            string
}

type VerifyDeviceChallengeCommand struct {
	TenantID           string
	CorrelationID      string
	ChallengeSessionID string
	Code               string
	TrustDevice        *bool
}

type QuickRevokeCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	Blacklist     bool
}

type QuickRevokeViewDTO struct {
	Message string         `json:"message,omitempty"`
	Payload map[string]any `json:"-"`
}

func QuickRevokeFromPayload(payload map[string]any) QuickRevokeViewDTO {
	view := QuickRevokeViewDTO{Payload: payload}
	if payload != nil {
		if msg, ok := payload["message"].(string); ok {
			view.Message = msg
		}
	}
	return view
}
