package client

import (
	"context"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
)

// CommunicationClient define a interface para comunicação com o ms-communication
type CommunicationClient interface {
	// SendMessage envia uma mensagem através do ms-communication
	SendMessage(ctx context.Context, req outboundDto.SendMessageRequestDTO, xApplication, correlationID string) (outboundDto.SendMessageResponseDTO, error)
}
