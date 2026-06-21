package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-auth/internal/infrastructure/validation"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestValidationMiddleware_GETRequest(t *testing.T) {
	// Arrange
	validator := validation.NewValidator()
	middleware := ValidationMiddleware(validator)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	// Act
	err := middleware(handler)(c)

	// Assert
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestValidationMiddleware_POSTRequest(t *testing.T) {
	// Arrange
	validator := validation.NewValidator()
	middleware := ValidationMiddleware(validator)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	// Act
	err := middleware(handler)(c)

	// Assert
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestValidationMiddleware_PUTRequest(t *testing.T) {
	// Arrange
	validator := validation.NewValidator()
	middleware := ValidationMiddleware(validator)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	// Act
	err := middleware(handler)(c)

	// Assert
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestValidateStruct_ValidStruct(t *testing.T) {
	// Arrange
	validator := validation.NewValidator()

	type TestStruct struct {
		Username string `validate:"required,username"`
		Email    string `validate:"required,email"`
		Password string `validate:"required,password"`
	}

	testStruct := TestStruct{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Act
	err := ValidateStruct(validator, testStruct)

	// Assert
	assert.NoError(t, err)
}

func TestValidateStruct_InvalidStruct(t *testing.T) {
	// Arrange
	validator := validation.NewValidator()

	type TestStruct struct {
		Username string `validate:"required,username"`
		Email    string `validate:"required,email"`
		Password string `validate:"required,password"`
	}

	testStruct := TestStruct{
		Username: "ab",            // Muito curto
		Email:    "invalid-email", // Email inválido
		Password: "123",           // Muito curto
	}

	// Act
	err := ValidateStruct(validator, testStruct)

	// Assert
	assert.Error(t, err)
}

func TestFormatValidationError_WithValidationErrors(t *testing.T) {
	// Arrange
	validator := validation.NewValidator()

	type TestStruct struct {
		Username string `validate:"required,username"`
		Email    string `validate:"required,email"`
		Password string `validate:"required,password"`
	}

	testStruct := TestStruct{
		Username: "ab",            // Muito curto
		Email:    "invalid-email", // Email inválido
		Password: "123",           // Muito curto
	}

	validationErr := ValidateStruct(validator, testStruct)

	// Act
	errorResponse := FormatValidationError(validationErr)

	// Assert
	assert.Equal(t, "VALIDATION_ERROR", errorResponse.Error)
	assert.Equal(t, "Dados de entrada inválidos", errorResponse.Message)
	assert.NotEmpty(t, errorResponse.Details)
	assert.GreaterOrEqual(t, len(errorResponse.Details), 3) // Pelo menos 3 erros de validação
}

func TestFormatValidationError_WithNilError(t *testing.T) {
	// Act
	errorResponse := FormatValidationError(nil)

	// Assert
	assert.Equal(t, "VALIDATION_ERROR", errorResponse.Error)
	assert.Equal(t, "Dados de entrada inválidos", errorResponse.Message)
	assert.Empty(t, errorResponse.Details)
}

func TestFormatValidationError_WithNonValidationError(t *testing.T) {
	// Arrange
	err := assert.AnError

	// Act
	errorResponse := FormatValidationError(err)

	// Assert
	assert.Equal(t, "VALIDATION_ERROR", errorResponse.Error)
	assert.Equal(t, "Dados de entrada inválidos", errorResponse.Message)
	assert.Len(t, errorResponse.Details, 1)
	assert.Equal(t, "unknown", errorResponse.Details[0].Field)
	assert.Equal(t, "unknown", errorResponse.Details[0].Code)
	assert.Equal(t, err.Error(), errorResponse.Details[0].Message)
}
