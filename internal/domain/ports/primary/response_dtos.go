package primary

// ErrorResponse representa uma resposta de erro padronizada
// @Description Resposta padronizada para erros da API
type ErrorResponse struct {
	// @Description Código do erro
	// @Example INVALID_REQUEST
	Error string `json:"error" example:"INVALID_REQUEST"`

	// @Description Mensagem descritiva do erro
	// @Example Requisição inválida
	Message string `json:"message" example:"Requisição inválida"`

	// @Description ID de rastreamento da requisição
	// @Example 12345678-1234-1234-1234-123456789abc
	TraceID string `json:"traceId,omitempty" example:"12345678-1234-1234-1234-123456789abc"`
}

// SuccessResponse representa uma resposta de sucesso padronizada
// @Description Resposta padronizada para operações bem-sucedidas
type SuccessResponse struct {
	// @Description Mensagem de sucesso
	// @Example Operação realizada com sucesso
	Message string `json:"message" example:"Operação realizada com sucesso"`

	// @Description ID de rastreamento da requisição
	// @Example 12345678-1234-1234-1234-123456789abc
	TraceID string `json:"traceId,omitempty" example:"12345678-1234-1234-1234-123456789abc"`
}
