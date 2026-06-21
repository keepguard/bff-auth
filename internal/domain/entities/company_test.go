package entities

import (
	"testing"

	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestNewCompany(t *testing.T) {
	tests := []struct {
		name         string
		id           valueobjects.CompanyID
		companyName  valueobjects.CompanyName
		xApplication valueobjects.XApplication
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid company",
			id:           func() valueobjects.CompanyID { id, _ := valueobjects.NewCompanyID("company-123"); return id }(),
			companyName:  func() valueobjects.CompanyName { name, _ := valueobjects.NewCompanyName("Test Company"); return name }(),
			xApplication: func() valueobjects.XApplication { app, _ := valueobjects.NewXApplication("test-app"); return app }(),
			wantErr:      false,
		},
		{
			name:         "empty company ID",
			id:           func() valueobjects.CompanyID { id, _ := valueobjects.NewCompanyID(""); return id }(),
			companyName:  func() valueobjects.CompanyName { name, _ := valueobjects.NewCompanyName("Test Company"); return name }(),
			xApplication: func() valueobjects.XApplication { app, _ := valueobjects.NewXApplication("test-app"); return app }(),
			wantErr:      true,
			errMsg:       "company ID cannot be empty",
		},
		{
			name:         "empty company name",
			id:           func() valueobjects.CompanyID { id, _ := valueobjects.NewCompanyID("company-123"); return id }(),
			companyName:  func() valueobjects.CompanyName { name, _ := valueobjects.NewCompanyName(""); return name }(),
			xApplication: func() valueobjects.XApplication { app, _ := valueobjects.NewXApplication("test-app"); return app }(),
			wantErr:      true,
			errMsg:       "company name cannot be empty",
		},
		{
			name:         "empty x-application",
			id:           func() valueobjects.CompanyID { id, _ := valueobjects.NewCompanyID("company-123"); return id }(),
			companyName:  func() valueobjects.CompanyName { name, _ := valueobjects.NewCompanyName("Test Company"); return name }(),
			xApplication: func() valueobjects.XApplication { app, _ := valueobjects.NewXApplication(""); return app }(),
			wantErr:      true,
			errMsg:       "x-application cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			company, err := NewCompany(tt.id, tt.companyName, tt.xApplication)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, company)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, company)
				assert.Equal(t, tt.id, company.ID())
				assert.Equal(t, tt.companyName, company.Name())
				assert.Equal(t, tt.xApplication, company.XApplication())
				assert.Equal(t, CompanyStatusActive, company.Status())
				assert.True(t, company.IsActive())
			}
		})
	}
}

func TestCompany_Activate(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	company.Deactivate() // Garante que está inativo
	
	company.Activate()
	
	assert.Equal(t, CompanyStatusActive, company.Status())
	assert.True(t, company.IsActive())
}

func TestCompany_Deactivate(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	company.Deactivate()
	
	assert.Equal(t, CompanyStatusInactive, company.Status())
	assert.False(t, company.IsActive())
}

func TestCompany_Block(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	company.Block()
	
	assert.Equal(t, CompanyStatusBlocked, company.Status())
	assert.True(t, company.IsBlocked())
}

func TestCompany_SetPending(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	company.SetPending()
	
	assert.Equal(t, CompanyStatusPending, company.Status())
	assert.True(t, company.IsPending())
}

func TestCompany_UpdateName(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	newName, _ := valueobjects.NewCompanyName("Updated Company")
	err := company.UpdateName(newName)
	
	assert.NoError(t, err)
	assert.Equal(t, newName, company.Name())
}

func TestCompany_UpdateName_Empty(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	emptyName, _ := valueobjects.NewCompanyName("")
	err := company.UpdateName(emptyName)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "company name cannot be empty")
}

func TestCompany_UpdateXApplication(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	newApp, _ := valueobjects.NewXApplication("new-app")
	err := company.UpdateXApplication(newApp)
	
	assert.NoError(t, err)
	assert.Equal(t, newApp, company.XApplication())
}

func TestCompany_UpdateXApplication_Empty(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	emptyApp, _ := valueobjects.NewXApplication("")
	err := company.UpdateXApplication(emptyApp)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "x-application cannot be empty")
}

func TestCompany_CanBeActivated(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	// Company ativa não pode ser ativada
	assert.False(t, company.CanBeActivated())
	
	// Company inativa pode ser ativada
	company.Deactivate()
	assert.True(t, company.CanBeActivated())
	
	// Company pendente pode ser ativada
	company.SetPending()
	assert.True(t, company.CanBeActivated())
}

func TestCompany_CanBeBlocked(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	// Company ativa pode ser bloqueada
	assert.True(t, company.CanBeBlocked())
	
	// Company bloqueada não pode ser bloqueada
	company.Block()
	assert.False(t, company.CanBeBlocked())
}

func TestCompany_CanAuthenticateUsers(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	// Company ativa pode autenticar usuários
	assert.True(t, company.CanAuthenticateUsers())
	
	// Company inativa não pode autenticar usuários
	company.Deactivate()
	assert.False(t, company.CanAuthenticateUsers())
}

func TestCompany_CanCreateUsers(t *testing.T) {
	id, _ := valueobjects.NewCompanyID("company-123")
	name, _ := valueobjects.NewCompanyName("Test Company")
	app, _ := valueobjects.NewXApplication("test-app")
	
	company, _ := NewCompany(id, name, app)
	
	// Company ativa pode criar usuários
	assert.True(t, company.CanCreateUsers())
	
	// Company inativa não pode criar usuários
	company.Deactivate()
	assert.False(t, company.CanCreateUsers())
}
