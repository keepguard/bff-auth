package dto

// MSAuthErrorResponse representa a resposta de erro do ms-auth (ProblemDetail)
type MSAuthErrorResponse struct {
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	Status  int                    `json:"status"`
	Detail  string                 `json:"detail"`
	Path    string                 `json:"path,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// HTTPError representa um erro HTTP genérico
type HTTPError struct {
	StatusCode int
	Message    string
	ErrorCode  string
	Details    map[string]interface{}
}

// Error implementa a interface error
func (e *HTTPError) Error() string {
	return e.Message
}
