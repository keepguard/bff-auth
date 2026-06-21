package dto

// SendMessageRequestDTO representa a requisição de envio de mensagem para o ms-communication
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

// SendMessageResponseDTO representa a resposta do ms-communication
type SendMessageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
