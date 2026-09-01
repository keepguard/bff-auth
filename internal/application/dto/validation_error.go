package dto

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string
	Message string
}

// Error implementa a interface error
func (e *ValidationError) Error() string {
	return e.Message
}

// IsValidationError verifica se o erro é do tipo ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
