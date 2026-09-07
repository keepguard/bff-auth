package dto

// Tipos de wire MS / contrato autenticado usados pelas portas de saída.
// JSON tags ficam aqui para o client HTTP; o adapter inbound pode aliasar estes tipos.

type AuthRequestDTO struct {
	Username  string `json:"username" validate:"required,username"`
	Password  string `json:"password" validate:"required,password"`
	CompanyID string `json:"companyId,omitempty"`
}

type RefreshTokenRequestDTO struct {
	Token string `json:"token" validate:"required,min=1"`
}

type AvailableMfaChannelDTO struct {
	Channel      string `json:"channel"`
	TargetMasked string `json:"targetMasked"`
	Description  string `json:"description"`
}

type AuthResponseDTO struct {
	Token              string                   `json:"token,omitempty"`
	ExpiresIn          int64                    `json:"expiresIn,omitempty"`
	Status             string                   `json:"status,omitempty"`
	ChallengeSessionID string                   `json:"challengeSessionId,omitempty"`
	IsTrusted          *bool                    `json:"isTrusted,omitempty"`
	AvailableChannels  []AvailableMfaChannelDTO `json:"availableChannels,omitempty"`
}

type RefreshTokenResponseDTO struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

type DeviceChallengeSendRequestDTO struct {
	ChallengeSessionID string `json:"challengeSessionId" validate:"required"`
	Channel            string `json:"channel" validate:"required"`
}

type DeviceChallengeVerifyRequestDTO struct {
	ChallengeSessionID string `json:"challengeSessionId" validate:"required"`
	Code               string `json:"code" validate:"required"`
	TrustDevice        *bool  `json:"trustDevice,omitempty"`
}

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

type DeviceBlacklistDTO struct {
	CodeUser   string `json:"codeUser"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	Reason     string `json:"reason"`
	BlockedAt  string `json:"blockedAt"`
	BlockedBy  string `json:"blockedBy"`
}

type AddDeviceBlacklistRequestDTO struct {
	DeviceID   string `json:"deviceId" validate:"required"`
	DeviceName string `json:"deviceName,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type AdminAddDeviceBlacklistRequestDTO struct {
	UserID     string `json:"userId" validate:"required"`
	DeviceID   string `json:"deviceId" validate:"required"`
	DeviceName string `json:"deviceName,omitempty"`
	Reason     string `json:"reason,omitempty"`
	ExpiresAt  string `json:"expiresAt,omitempty"`
}

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

type PaginatedDeviceSessionResponseDTO struct {
	Content          []DeviceSessionDTO `json:"content"`
	Page             int                `json:"page"`
	Number           int                `json:"number"`
	Size             int                `json:"size"`
	TotalElements    int64              `json:"totalElements"`
	TotalPages       int                `json:"totalPages"`
	First            bool               `json:"first"`
	Last             bool               `json:"last"`
	NumberOfElements int                `json:"numberOfElements"`
	Empty            bool               `json:"empty"`
}

type ChangePasswordMSRequestDTO struct {
	CodeUser           string `json:"codeUser"`
	CurrentPassword    string `json:"currentPassword"`
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
}

type ResetPasswordMSRequestDTO struct {
	CodeUser           string `json:"codeUser"`
	ResetToken         string `json:"resetToken"`
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
	MessageType        string `json:"messageType"`
	TemplateType       string `json:"templateType"`
}

type UserByEmailResponseDTO struct {
	ID            string `json:"id"`
	CodeUser      string `json:"codeUser"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"emailVerified"`
}

type UserByCodeResponseDTO struct {
	ID                string `json:"id"`
	IDUserExternal    string `json:"idUserExternal"`
	IDUserExternalAlt string `json:"id_user_external"`
	CodeUser          string `json:"codeUser"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Status            string `json:"status"`
	CompanyID         string `json:"companyId"`
}

func (u UserByCodeResponseDTO) ExternalID() string {
	if u.IDUserExternal != "" {
		return u.IDUserExternal
	}
	return u.IDUserExternalAlt
}

type GenerateResetTokenMSRequestDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
}

type GenerateResetTokenMSResponseDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
	Token             string `json:"token"`
	ExpiresInSeconds  int64  `json:"expiresInSeconds"`
}

type SendMessageRequestDTO struct {
	MessageType       string                 `json:"messageType"`
	CommunicationType string                 `json:"communicationType"`
	TemplateType      string                 `json:"templateType"`
	Recipient         string                 `json:"recipient"`
	Subject           string                 `json:"subject,omitempty"`
	Content           string                 `json:"content,omitempty"`
	CodeUser          string                 `json:"codeUser,omitempty"`
	Variables         map[string]interface{} `json:"variables"`
}

type SendMessageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CompanySimpleResponseDTO struct {
	ID        string `json:"id"`
	TenantId  string `json:"tenantId"`
	Name      string `json:"name"`
	LegalName string `json:"legalName"`
	CNPJ      string `json:"cnpj"`
	Status    string `json:"status"`
}
