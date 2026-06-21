package entities

import (
	"time"

	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
)

// User representa a entidade de usuário no domínio
type User struct {
	id           valueobjects.UserID
	username     valueobjects.Username
	email        valueobjects.Email
	active       bool
	roles        []valueobjects.Role
	applications []valueobjects.Application
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUser cria uma nova instância de usuário
func NewUser(id valueobjects.UserID, username valueobjects.Username, email valueobjects.Email) *User {
	now := time.Now()
	return &User{
		id:           id,
		username:     username,
		email:        email,
		active:       true,
		roles:        []valueobjects.Role{},
		applications: []valueobjects.Application{},
		createdAt:    now,
		updatedAt:    now,
	}
}

// Getters
func (u *User) ID() valueobjects.UserID {
	return u.id
}

func (u *User) Username() valueobjects.Username {
	return u.username
}

func (u *User) Email() valueobjects.Email {
	return u.email
}

func (u *User) IsActive() bool {
	return u.active
}

func (u *User) Roles() []valueobjects.Role {
	return u.roles
}

func (u *User) Applications() []valueobjects.Application {
	return u.applications
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Business Methods
func (u *User) AddRole(role valueobjects.Role) {
	// Verifica se o role já existe
	for _, r := range u.roles {
		if r == role {
			return
		}
	}
	u.roles = append(u.roles, role)
	u.updatedAt = time.Now()
}

func (u *User) RemoveRole(role valueobjects.Role) {
	for i, r := range u.roles {
		if r == role {
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			u.updatedAt = time.Now()
			return
		}
	}
}

func (u *User) AddApplication(app valueobjects.Application) {
	// Verifica se a aplicação já existe
	for _, a := range u.applications {
		if a == app {
			return
		}
	}
	u.applications = append(u.applications, app)
	u.updatedAt = time.Now()
}

func (u *User) Activate() {
	u.active = true
	u.updatedAt = time.Now()
}

func (u *User) Deactivate() {
	u.active = false
	u.updatedAt = time.Now()
}

func (u *User) HasRole(role valueobjects.Role) bool {
	for _, r := range u.roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) HasApplication(app valueobjects.Application) bool {
	for _, a := range u.applications {
		if a == app {
			return true
		}
	}
	return false
}
