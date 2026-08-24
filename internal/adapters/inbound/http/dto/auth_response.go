package dto

// AvailableMfaChannelDTO representa um canal de MFA disponível para escolha
type AvailableMfaChannelDTO struct {
	Channel      string `json:"channel"`
	TargetMasked string `json:"targetMasked"`
	Description  string `json:"description"`
}

// AuthResponseDTO representa a resposta de autenticação
type AuthResponseDTO struct {
	Token              string                   `json:"token,omitempty"`
	ExpiresIn          int64                    `json:"expiresIn,omitempty"`
	Status             string                   `json:"status,omitempty"`
	ChallengeSessionID string                   `json:"challengeSessionId,omitempty"`
	IsTrusted          *bool                    `json:"isTrusted,omitempty"`
	AvailableChannels  []AvailableMfaChannelDTO `json:"availableChannels,omitempty"`
}

// RefreshTokenResponseDTO representa a resposta de refresh de token
type RefreshTokenResponseDTO struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

// DeviceChallengeSendRequestDTO representa a requisição para disparo de OTP de dispositivo
type DeviceChallengeSendRequestDTO struct {
	ChallengeSessionID string `json:"challengeSessionId" validate:"required"`
	Channel            string `json:"channel" validate:"required"`
}

// DeviceChallengeVerifyRequestDTO representa a requisição para validação de OTP de dispositivo
type DeviceChallengeVerifyRequestDTO struct {
	ChallengeSessionID string `json:"challengeSessionId" validate:"required"`
	Code               string `json:"code" validate:"required"`
	TrustDevice        *bool  `json:"trustDevice,omitempty"`
}

// DeviceSessionDTO representa uma sessão ativa por dispositivo
type DeviceSessionDTO struct {
	SessionID    string `json:"sessionId"`
	DeviceID     string `json:"deviceId"`
	DeviceName   string `json:"deviceName"`
	DeviceType   string `json:"deviceType"`
	IPAddress    string `json:"ipAddress"`
	Location     string `json:"location"`
	IsCurrent    bool   `json:"isCurrent"`
	IsTrusted    bool   `json:"isTrusted"`
	LastActiveAt string `json:"lastActiveAt"`
	CreatedAt    string `json:"createdAt"`
}

// DeviceBlacklistDTO representa um dispositivo na blacklist
type DeviceBlacklistDTO struct {
	CodeUser   string `json:"codeUser"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	Reason     string `json:"reason"`
	BlockedAt  string `json:"blockedAt"`
	BlockedBy  string `json:"blockedBy"`
}

// AddDeviceBlacklistRequestDTO representa a requisição para adicionar dispositivo à blacklist
type AddDeviceBlacklistRequestDTO struct {
	DeviceID   string `json:"deviceId" validate:"required"`
	DeviceName string `json:"deviceName,omitempty"`
	Reason     string `json:"reason,omitempty"`
}
