package entities

import (
	"errors"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
)

// Permission representa uma permissão no sistema
type Permission struct {
	id       valueobjects.PermissionID
	name     valueobjects.PermissionName
	resource valueobjects.Resource
	action   valueobjects.Action
	scope    valueobjects.Scope
	active   bool
	createdAt time.Time
	updatedAt time.Time
}

// NewPermission cria uma nova permissão
func NewPermission(
	id valueobjects.PermissionID,
	name valueobjects.PermissionName,
	resource valueobjects.Resource,
	action valueobjects.Action,
	scope valueobjects.Scope,
) (*Permission, error) {
	if id.IsEmpty() {
		return nil, errors.New("permission ID cannot be empty")
	}
	
	if name.IsEmpty() {
		return nil, errors.New("permission name cannot be empty")
	}
	
	if resource.IsEmpty() {
		return nil, errors.New("resource cannot be empty")
	}
	
	if action.IsEmpty() {
		return nil, errors.New("action cannot be empty")
	}
	
	if scope.IsEmpty() {
		return nil, errors.New("scope cannot be empty")
	}
	
	now := time.Now()
	return &Permission{
		id:       id,
		name:     name,
		resource: resource,
		action:   action,
		scope:    scope,
		active:   true,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Getters
func (p *Permission) ID() valueobjects.PermissionID {
	return p.id
}

func (p *Permission) Name() valueobjects.PermissionName {
	return p.name
}

func (p *Permission) Resource() valueobjects.Resource {
	return p.resource
}

func (p *Permission) Action() valueobjects.Action {
	return p.action
}

func (p *Permission) Scope() valueobjects.Scope {
	return p.scope
}

func (p *Permission) IsActive() bool {
	return p.active
}

func (p *Permission) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Permission) UpdatedAt() time.Time {
	return p.updatedAt
}

// Business Methods
func (p *Permission) Activate() {
	p.active = true
	p.updatedAt = time.Now()
}

func (p *Permission) Deactivate() {
	p.active = false
	p.updatedAt = time.Now()
}

func (p *Permission) UpdateName(name valueobjects.PermissionName) error {
	if name.IsEmpty() {
		return errors.New("permission name cannot be empty")
	}
	p.name = name
	p.updatedAt = time.Now()
	return nil
}

func (p *Permission) UpdateResource(resource valueobjects.Resource) error {
	if resource.IsEmpty() {
		return errors.New("resource cannot be empty")
	}
	p.resource = resource
	p.updatedAt = time.Now()
	return nil
}

func (p *Permission) UpdateAction(action valueobjects.Action) error {
	if action.IsEmpty() {
		return errors.New("action cannot be empty")
	}
	p.action = action
	p.updatedAt = time.Now()
	return nil
}

func (p *Permission) UpdateScope(scope valueobjects.Scope) error {
	if scope.IsEmpty() {
		return errors.New("scope cannot be empty")
	}
	p.scope = scope
	p.updatedAt = time.Now()
	return nil
}

func (p *Permission) Matches(resource valueobjects.Resource, action valueobjects.Action) bool {
	return p.active && p.resource.Equals(resource) && p.action.Equals(action)
}

func (p *Permission) HasScope(scope valueobjects.Scope) bool {
	return p.active && p.scope.Equals(scope)
}

func (p *Permission) CanAccess(resource valueobjects.Resource, action valueobjects.Action, scope valueobjects.Scope) bool {
	return p.active && 
		   p.resource.Equals(resource) && 
		   p.action.Equals(action) && 
		   p.scope.Equals(scope)
}

func (p *Permission) IsSystemPermission() bool {
	return p.scope.Equals(valueobjects.ScopeSystem)
}

func (p *Permission) IsApplicationPermission() bool {
	return p.scope.Equals(valueobjects.ScopeApplication)
}

func (p *Permission) IsUserPermission() bool {
	return p.scope.Equals(valueobjects.ScopeUser)
}
