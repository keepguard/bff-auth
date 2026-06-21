package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	// Arrange
	err := &ValidationError{
		Field:   "username",
		Message: "username é obrigatório",
	}

	// Act
	result := err.Error()

	// Assert
	assert.Equal(t, "username é obrigatório", result)
}

func TestServiceUnavailableError_Error(t *testing.T) {
	// Arrange
	err := &ServiceUnavailableError{
		Service: "ms-auth",
		Message: "Serviço de autenticação indisponível",
		Details: map[string]interface{}{
			"timeout": "30s",
			"retries": 3,
		},
	}

	// Act
	result := err.Error()

	// Assert
	assert.Equal(t, "Serviço de autenticação indisponível", result)
}

func TestValidationError_WithEmptyFields(t *testing.T) {
	// Arrange
	err := &ValidationError{
		Field:   "",
		Message: "",
	}

	// Act
	result := err.Error()

	// Assert
	assert.Equal(t, "", result)
}

func TestServiceUnavailableError_WithEmptyMessage(t *testing.T) {
	// Arrange
	err := &ServiceUnavailableError{
		Service: "ms-company",
		Message: "",
		Details: nil,
	}

	// Act
	result := err.Error()

	// Assert
	assert.Equal(t, "", result)
}

func TestValidationError_AsError(t *testing.T) {
	// Arrange
	var err error = &ValidationError{
		Field:   "email",
		Message: "email inválido",
	}

	// Act & Assert
	assert.Error(t, err)
	assert.Equal(t, "email inválido", err.Error())
}

func TestServiceUnavailableError_AsError(t *testing.T) {
	// Arrange
	var err error = &ServiceUnavailableError{
		Service: "ms-user",
		Message: "Timeout ao conectar com ms-user",
		Details: map[string]interface{}{
			"status": "circuit breaker open",
		},
	}

	// Act & Assert
	assert.Error(t, err)
	assert.Equal(t, "Timeout ao conectar com ms-user", err.Error())
}
