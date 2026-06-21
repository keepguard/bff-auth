package dto

// AuthResponseDTO representa a resposta de autenticação
type AuthResponseDTO struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

// RefreshTokenResponseDTO representa a resposta de refresh de token
type RefreshTokenResponseDTO struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}
