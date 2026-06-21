package swagger

import (
	"github.com/swaggo/swag"
)

// InitSwagger inicializa a configuração do Swagger
func InitSwagger() {
	swag.Register("swagger", &swag.Spec{
		Version:          "1.0.0",
		Host:             "localhost:8381",
		BasePath:         "/",
		Schemes:          []string{},
		Title:            "BFF-AUTH API",
		Description:      "Backend for Frontend responsável pela autenticação e autorização do sistema KeepGuard",
		InfoInstanceName: "swagger",
		SwaggerTemplate:  getSwaggerTemplate(),
		LeftDelim:        "{{",
		RightDelim:       "}}",
	})
}

// getSwaggerTemplate retorna o template base do Swagger
func getSwaggerTemplate() string {
	return `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {},
    "definitions": {},
    "securityDefinitions": {
        "Bearer": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`
}
