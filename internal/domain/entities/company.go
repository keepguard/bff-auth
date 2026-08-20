package entities

import (
	"errors"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
)

// Company representa uma empresa no domínio de autenticação
type Company struct {
	id           valueobjects.CompanyID
	name         valueobjects.CompanyName
	tenantId valueobjects.TenantId
	status       CompanyStatus
	createdAt    time.Time
	updatedAt    time.Time
}

// CompanyStatus representa o status da empresa
type CompanyStatus string

const (
	CompanyStatusActive   CompanyStatus = "ACTIVE"
	CompanyStatusInactive CompanyStatus = "INACTIVE"
	CompanyStatusPending  CompanyStatus = "PENDING"
	CompanyStatusBlocked  CompanyStatus = "BLOCKED"
)

// NewCompany cria uma nova empresa
func NewCompany(id valueobjects.CompanyID, name valueobjects.CompanyName, tenantId valueobjects.TenantId) (*Company, error) {
	if id.IsEmpty() {
		return nil, errors.New("company ID cannot be empty")
	}
	
	if name.IsEmpty() {
		return nil, errors.New("company name cannot be empty")
	}
	
	if tenantId.IsEmpty() {
		return nil, errors.New("x-tenant-id cannot be empty")
	}
	
	now := time.Now()
	return &Company{
		id:           id,
		name:         name,
		tenantId: tenantId,
		status:       CompanyStatusActive,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// Getters
func (c *Company) ID() valueobjects.CompanyID {
	return c.id
}

func (c *Company) Name() valueobjects.CompanyName {
	return c.name
}

func (c *Company) TenantId() valueobjects.TenantId {
	return c.tenantId
}

func (c *Company) Status() CompanyStatus {
	return c.status
}

func (c *Company) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Company) UpdatedAt() time.Time {
	return c.updatedAt
}

// Business Methods
func (c *Company) Activate() {
	c.status = CompanyStatusActive
	c.updatedAt = time.Now()
}

func (c *Company) Deactivate() {
	c.status = CompanyStatusInactive
	c.updatedAt = time.Now()
}

func (c *Company) Block() {
	c.status = CompanyStatusBlocked
	c.updatedAt = time.Now()
}

func (c *Company) SetPending() {
	c.status = CompanyStatusPending
	c.updatedAt = time.Now()
}

func (c *Company) UpdateName(name valueobjects.CompanyName) error {
	if name.IsEmpty() {
		return errors.New("company name cannot be empty")
	}
	c.name = name
	c.updatedAt = time.Now()
	return nil
}

func (c *Company) UpdateTenantId(tenantId valueobjects.TenantId) error {
	if tenantId.IsEmpty() {
		return errors.New("x-tenant-id cannot be empty")
	}
	c.tenantId = tenantId
	c.updatedAt = time.Now()
	return nil
}

func (c *Company) IsActive() bool {
	return c.status == CompanyStatusActive
}

func (c *Company) IsBlocked() bool {
	return c.status == CompanyStatusBlocked
}

func (c *Company) IsPending() bool {
	return c.status == CompanyStatusPending
}

func (c *Company) CanBeActivated() bool {
	return c.status == CompanyStatusInactive || c.status == CompanyStatusPending
}

func (c *Company) CanBeBlocked() bool {
	return c.status == CompanyStatusActive
}

func (c *Company) CanAuthenticateUsers() bool {
	return c.status == CompanyStatusActive
}

func (c *Company) CanCreateUsers() bool {
	return c.status == CompanyStatusActive
}
