package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCompanyClient_GetByTenantId_Success(t *testing.T) {
	// Arrange
	expectedResponse := client.CompanySimpleResponseDTO{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		TenantId: "550e8400-e29b-41d4-a716-446655440000",
		Name:         "Empresa Teste",
		LegalName:    "Empresa Teste LTDA",
		CNPJ:         "12345678000199",
		Status:       "ACTIVE",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/companies/x-tenant-id/550e8400-e29b-41d4-a716-446655440000", r.URL.Path)
		assert.Equal(t, "test-correlation-id", r.Header.Get("X-Correlation-ID"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Company: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}

	logger, _ := zap.NewDevelopment()
	companyClient := NewCompanyClient(cfg, logger)

	// Act
	result, err := companyClient.GetByTenantId(
		context.Background(),
		"550e8400-e29b-41d4-a716-446655440000",
		"test-correlation-id",
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
}

func TestCompanyClient_GetByTenantId_NotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Empresa não encontrada"}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Company: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}

	logger, _ := zap.NewDevelopment()
	companyClient := NewCompanyClient(cfg, logger)

	// Act
	result, err := companyClient.GetByTenantId(
		context.Background(),
		"550e8400-e29b-41d4-a716-446655440000",
		"test-correlation-id",
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Empresa não encontrada")
	assert.Equal(t, client.CompanySimpleResponseDTO{}, result)
}

func TestCompanyClient_GetByTenantId_ServerError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Erro interno do servidor"}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Company: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}

	logger, _ := zap.NewDevelopment()
	companyClient := NewCompanyClient(cfg, logger)

	// Act
	result, err := companyClient.GetByTenantId(
		context.Background(),
		"550e8400-e29b-41d4-a716-446655440000",
		"test-correlation-id",
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Erro interno do servidor")
	assert.Equal(t, client.CompanySimpleResponseDTO{}, result)
}

func TestCompanyClient_GetByTenantId_InvalidJSON(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	cfg := &config.Config{
		Services: config.ServicesConfig{
			Company: config.ServiceConfig{
				BaseURL: server.URL,
			},
		},
	}

	logger, _ := zap.NewDevelopment()
	companyClient := NewCompanyClient(cfg, logger)

	// Act
	result, err := companyClient.GetByTenantId(
		context.Background(),
		"550e8400-e29b-41d4-a716-446655440000",
		"test-correlation-id",
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao fazer parse da resposta")
	assert.Equal(t, client.CompanySimpleResponseDTO{}, result)
}
