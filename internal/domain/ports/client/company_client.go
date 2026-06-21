package client

import (
	"context"
)

// CompanySimpleResponseDTO representa a resposta simples da empresa
type CompanySimpleResponseDTO struct {
	ID           string `json:"id"`
	XApplication string `json:"xApplication"`
	Name         string `json:"name"`
	LegalName    string `json:"legalName"`
	CNPJ         string `json:"cnpj"`
	Status       string `json:"status"`
}

// CompanyClient interface para comunicação com o serviço de empresas
type CompanyClient interface {
	// GetByXApplication busca uma empresa pelo X-Application
	GetByXApplication(ctx context.Context, xApplication, correlationID string) (CompanySimpleResponseDTO, error)
}
