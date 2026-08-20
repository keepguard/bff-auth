package client

import (
	"context"
)

// CompanySimpleResponseDTO representa a resposta simples da empresa
type CompanySimpleResponseDTO struct {
	ID           string `json:"id"`
	TenantId string `json:"tenantId"`
	Name         string `json:"name"`
	LegalName    string `json:"legalName"`
	CNPJ         string `json:"cnpj"`
	Status       string `json:"status"`
}

// CompanyClient interface para comunicação com o serviço de empresas
type CompanyClient interface {
	// GetByTenantId busca uma empresa pelo X-Tenant-Id
	GetByTenantId(ctx context.Context, tenantId, correlationID string) (CompanySimpleResponseDTO, error)
}
