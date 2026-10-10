package dto

import (
	"encoding/json"
	"testing"
)

// Contrato com o ms-company-go: a resposta (e o JSON do cache Redis) ganhou
// documentType, document e product, e a empresa pode ser PF (cnpj "",
// taxRegime PESSOA_FISICA, datas null). O decoder ignora campo novo.
func TestCompanySimpleResponseDecodesIndividualCompany(t *testing.T) {
	raw := `{"id":"2b7e1c1e-7a52-4f5b-9a3e-1f2d3c4b5a69","codeCompany":"f7fc7350-b9fc-4e54-9c58-ac9385b23ae4",
		"tenantId":"0d9a4c3e-2b1a-4f6e-8d7c-5b4a3c2d1e0f","name":"Achadinhos","legalName":"Maria da Silva",
		"cnpj":"","documentType":"CPF","document":"12345678909","product":"achadinhos",
		"taxRegime":"PESSOA_FISICA","mfaChannels":[],"status":"ACTIVE","createdAt":null,"updatedAt":null}`

	var c CompanySimpleResponseDTO
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatalf("empresa PF não decodificou: %v", err)
	}
	if c.TenantId != "0d9a4c3e-2b1a-4f6e-8d7c-5b4a3c2d1e0f" || c.Status != "ACTIVE" || c.Name != "Achadinhos" || c.CNPJ != "" {
		t.Fatalf("campos errados: %+v", c)
	}
}
