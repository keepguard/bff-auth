package http

import (
	"github.com/labstack/echo/v4"
)

// CombinedHandlers combina AuthHandlers e MessageHandlers
type CombinedHandlers struct {
	*AuthHandlers
	*MessageHandlers
}

// NewCombinedHandlers cria uma nova instância de CombinedHandlers
func NewCombinedHandlers(
	authHandlers *AuthHandlers,
	messageHandlers *MessageHandlers,
) *CombinedHandlers {
	return &CombinedHandlers{
		AuthHandlers:    authHandlers,
		MessageHandlers: messageHandlers,
	}
}

// Implementa todos os métodos da interface Handler
func (h *CombinedHandlers) LoginHandler(c echo.Context) error {
	return h.AuthHandlers.LoginHandler(c)
}

func (h *CombinedHandlers) RefreshHandler(c echo.Context) error {
	return h.AuthHandlers.RefreshHandler(c)
}

func (h *CombinedHandlers) LogoutHandler(c echo.Context) error {
	return h.AuthHandlers.LogoutHandler(c)
}

func (h *CombinedHandlers) ValidateTokenHandler(c echo.Context) error {
	return h.AuthHandlers.ValidateTokenHandler(c)
}

func (h *CombinedHandlers) ChangePasswordHandler(c echo.Context) error {
	return h.AuthHandlers.ChangePasswordHandler(c)
}

func (h *CombinedHandlers) ResetPasswordHandler(c echo.Context) error {
	return h.AuthHandlers.ResetPasswordHandler(c)
}

func (h *CombinedHandlers) SendResetPasswordMessageHandler(c echo.Context) error {
	return h.MessageHandlers.SendResetPasswordMessageHandler(c)
}

func (h *CombinedHandlers) SendDeviceChallengeHandler(c echo.Context) error {
	return h.AuthHandlers.SendDeviceChallengeHandler(c)
}

func (h *CombinedHandlers) VerifyDeviceChallengeHandler(c echo.Context) error {
	return h.AuthHandlers.VerifyDeviceChallengeHandler(c)
}

func (h *CombinedHandlers) ListUserSessionsHandler(c echo.Context) error {
	return h.AuthHandlers.ListUserSessionsHandler(c)
}

func (h *CombinedHandlers) RevokeSessionHandler(c echo.Context) error {
	return h.AuthHandlers.RevokeSessionHandler(c)
}

func (h *CombinedHandlers) RevokeAllOtherSessionsHandler(c echo.Context) error {
	return h.AuthHandlers.RevokeAllOtherSessionsHandler(c)
}
