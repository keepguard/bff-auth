package http

import (
	"github.com/keepguard/bff-auth/internal/infrastructure/validation"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
)

// ValidationMiddleware cria um middleware de validação
func ValidationMiddleware(validator validation.Validator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Apenas valida requisições POST e PUT
			if c.Request().Method != "POST" && c.Request().Method != "PUT" {
				return next(c)
			}

			// Para simplificar, não valida por rota específica
			// A validação será feita nos handlers individuais
			return next(c)
		}
	}
}

// ValidateStruct valida uma struct usando o validador
func ValidateStruct(validator validation.Validator, s interface{}) error {
	return validator.ValidateStruct(s)
}

// FormatValidationError formata um erro de validação para resposta HTTP
func FormatValidationError(err error) pkg.ErrorResponse {
	validationErrors := validation.FormatValidationErrors(err)

	var details []pkg.ErrorDetail
	for _, ve := range validationErrors.Errors {
		details = append(details, pkg.ErrorDetail{
			Field:   ve.Field,
			Code:    ve.Tag,
			Message: ve.Message,
		})
	}

	return pkg.ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "Dados de entrada inválidos",
		Details: details,
	}
}
