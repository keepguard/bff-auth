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
	CodeUser     string `json:"codeUser,omitempty"`
	Writable     *bool  `json:"writable,omitempty"`
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

// AdminAddDeviceBlacklistRequestDTO representa a requisição administrativa para bloquear dispositivo
type AdminAddDeviceBlacklistRequestDTO struct {
	UserID     string `json:"userId" validate:"required"`
	DeviceID   string `json:"deviceId" validate:"required"`
	DeviceName string `json:"deviceName,omitempty"`
	Reason     string `json:"reason,omitempty"`
	ExpiresAt  string `json:"expiresAt,omitempty"`
}

// AdminDeviceBlacklistEntryDTO representa a entrada de blacklist retornada para admin
type AdminDeviceBlacklistEntryDTO struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenantId"`
	CodeUser   string `json:"codeUser"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	IPAddress  string `json:"ipAddress"`
	UserAgent  string `json:"userAgent"`
	Reason     string `json:"reason"`
	BlockedBy  string `json:"blockedBy"`
	BlockedAt  string `json:"blockedAt"`
	ExpiresAt  string `json:"expiresAt,omitempty"`
	Writable   *bool  `json:"writable,omitempty"`
}

// PaginatedDeviceBlacklistResponseDTO representa a resposta paginada de blacklist
type PaginatedDeviceBlacklistResponseDTO struct {
	Content          []AdminDeviceBlacklistEntryDTO `json:"content"`
	Page             int                            `json:"page"`
	Number           int                            `json:"number"`
	Size             int                            `json:"size"`
	TotalElements    int64                          `json:"totalElements"`
	TotalPages       int                            `json:"totalPages"`
	First            bool                           `json:"first"`
	Last             bool                           `json:"last"`
	NumberOfElements int                            `json:"numberOfElements"`
	Empty            bool                           `json:"empty"`
}

// PaginatedDeviceSessionResponseDTO representa a resposta paginada de sessões do tenant
type PaginatedDeviceSessionResponseDTO struct {
	Content          []DeviceSessionDTO `json:"content"`
	Page             int                 `json:"page"`
	Number           int                 `json:"number"`
	Size             int                 `json:"size"`
	TotalElements    int64              `json:"totalElements"`
	TotalPages       int                 `json:"totalPages"`
	First            bool               `json:"first"`
	Last             bool               `json:"last"`
	NumberOfElements int                 `json:"numberOfElements"`
	Empty            bool               `json:"empty"`
}
