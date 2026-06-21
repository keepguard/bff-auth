package services

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/entities"
	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
)

// UserDomainService define as regras de negócio para usuários
type UserDomainService interface {
	ValidateUserCreation(user entities.User) error
	ValidatePassword(password string) error
	CanUserBeActivated(user entities.User) error
	CanUserBeDeactivated(user entities.User) error
	GenerateUsername(email valueobjects.Email) valueobjects.Username
	ValidateUsername(username valueobjects.Username) error
	CanUserAddRole(user entities.User, role valueobjects.Role) error
	CanUserRemoveRole(user entities.User, role valueobjects.Role) error
	CanUserAddApplication(user entities.User, app valueobjects.Application) error
	CanUserRemoveApplication(user entities.User, app valueobjects.Application) error
	ValidateUserUpdate(user entities.User, newEmail valueobjects.Email, newUsername valueobjects.Username) error
}

type userDomainServiceImpl struct{}

// NewUserDomainService cria um novo serviço de domínio para usuários
func NewUserDomainService() UserDomainService {
	return &userDomainServiceImpl{}
}

// ValidateUserCreation valida se um usuário pode ser criado
func (s *userDomainServiceImpl) ValidateUserCreation(user entities.User) error {
	// Verificar se o usuário tem dados válidos
	if user.ID().IsEmpty() {
		return errors.New("user ID is required")
	}
	
	if user.Username().IsEmpty() {
		return errors.New("username is required")
	}
	
	if user.Email().IsEmpty() {
		return errors.New("email is required")
	}
	
	// Verificar se o usuário está ativo por padrão
	if !user.IsActive() {
		return errors.New("user must be active when created")
	}
	
	return nil
}

// ValidatePassword valida uma senha
func (s *userDomainServiceImpl) ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	
	if len(password) < 8 {
		return errors.New("password must have at least 8 characters")
	}
	
	if len(password) > 128 {
		return errors.New("password must have at most 128 characters")
	}
	
	// Verificar se contém pelo menos uma letra minúscula
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	
	// Verificar se contém pelo menos uma letra maiúscula
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	
	// Verificar se contém pelo menos um número
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	
	// Verificar se contém pelo menos um caractere especial
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	
	return nil
}

// CanUserBeActivated verifica se um usuário pode ser ativado
func (s *userDomainServiceImpl) CanUserBeActivated(user entities.User) error {
	if user.IsActive() {
		return errors.New("user is already active")
	}
	
	// Verificar se o usuário tem dados válidos
	if user.Username().IsEmpty() {
		return errors.New("cannot activate user without username")
	}
	
	if user.Email().IsEmpty() {
		return errors.New("cannot activate user without email")
	}
	
	return nil
}

// CanUserBeDeactivated verifica se um usuário pode ser desativado
func (s *userDomainServiceImpl) CanUserBeDeactivated(user entities.User) error {
	if !user.IsActive() {
		return errors.New("user is already inactive")
	}
	
	// Verificar se o usuário não tem roles críticos
	criticalRoles := []valueobjects.Role{
		valueobjects.RoleAdmin,
		valueobjects.RoleManager,
	}
	
	for _, role := range user.Roles() {
		for _, criticalRole := range criticalRoles {
			if role.Equals(criticalRole) {
				return errors.New("cannot deactivate user with critical role")
			}
		}
	}
	
	return nil
}

// GenerateUsername gera um username baseado no email
func (s *userDomainServiceImpl) GenerateUsername(email valueobjects.Email) valueobjects.Username {
	// Extrai a parte local do email e remove caracteres especiais
	localPart := email.LocalPart()
	username := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(localPart, "")
	
	// Garante que tenha pelo menos 3 caracteres
	if len(username) < 3 {
		username = username + "user"
	}
	
	// Limita a 50 caracteres
	if len(username) > 50 {
		username = username[:50]
	}
	
	// Adiciona timestamp para garantir unicidade
	timestamp := time.Now().Format("20060102150405")
	username = username + "_" + timestamp[len(timestamp)-6:] // últimos 6 dígitos
	
	// Cria o Username usando o construtor (pode retornar erro, mas assumimos que é válido)
	usernameVO, _ := valueobjects.NewUsername(username)
	return usernameVO
}

// ValidateUsername valida um username
func (s *userDomainServiceImpl) ValidateUsername(username valueobjects.Username) error {
	if username.IsEmpty() {
		return errors.New("username cannot be empty")
	}
	
	// Verificar se não contém caracteres inválidos
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username.Value()) {
		return errors.New("username can only contain letters, numbers, underscore and hyphen")
	}
	
	// Verificar se não começa ou termina com underscore ou hífen
	if strings.HasPrefix(username.Value(), "_") || strings.HasPrefix(username.Value(), "-") {
		return errors.New("username cannot start with underscore or hyphen")
	}
	
	if strings.HasSuffix(username.Value(), "_") || strings.HasSuffix(username.Value(), "-") {
		return errors.New("username cannot end with underscore or hyphen")
	}
	
	return nil
}

// CanUserAddRole verifica se um usuário pode adicionar um role
func (s *userDomainServiceImpl) CanUserAddRole(user entities.User, role valueobjects.Role) error {
	if !user.IsActive() {
		return errors.New("cannot add role to inactive user")
	}
	
	// Verificar se o role já existe
	for _, existingRole := range user.Roles() {
		if existingRole.Equals(role) {
			return errors.New("user already has this role")
		}
	}
	
	// Verificar se o usuário não está tentando adicionar um role de admin sem ter role de manager primeiro
	if role.Equals(valueobjects.RoleAdmin) {
		hasManager := false
		for _, existingRole := range user.Roles() {
			if existingRole.Equals(valueobjects.RoleManager) {
				hasManager = true
				break
			}
		}
		if !hasManager {
			return errors.New("user must have manager role before admin role")
		}
	}
	
	return nil
}

// CanUserRemoveRole verifica se um usuário pode remover um role
func (s *userDomainServiceImpl) CanUserRemoveRole(user entities.User, role valueobjects.Role) error {
	if !user.IsActive() {
		return errors.New("cannot remove role from inactive user")
	}
	
	// Verificar se o role existe
	hasRole := false
	for _, existingRole := range user.Roles() {
		if existingRole.Equals(role) {
			hasRole = true
			break
		}
	}
	
	if !hasRole {
		return errors.New("user does not have this role")
	}
	
	// Verificar se não está removendo o último role
	if len(user.Roles()) == 1 {
		return errors.New("user must have at least one role")
	}
	
	return nil
}

// CanUserAddApplication verifica se um usuário pode adicionar uma aplicação
func (s *userDomainServiceImpl) CanUserAddApplication(user entities.User, app valueobjects.Application) error {
	if !user.IsActive() {
		return errors.New("cannot add application to inactive user")
	}
	
	// Verificar se a aplicação já existe
	for _, existingApp := range user.Applications() {
		if existingApp.Equals(app) {
			return errors.New("user already has access to this application")
		}
	}
	
	return nil
}

// CanUserRemoveApplication verifica se um usuário pode remover uma aplicação
func (s *userDomainServiceImpl) CanUserRemoveApplication(user entities.User, app valueobjects.Application) error {
	if !user.IsActive() {
		return errors.New("cannot remove application from inactive user")
	}
	
	// Verificar se a aplicação existe
	hasApp := false
	for _, existingApp := range user.Applications() {
		if existingApp.Equals(app) {
			hasApp = true
			break
		}
	}
	
	if !hasApp {
		return errors.New("user does not have access to this application")
	}
	
	// Verificar se não está removendo a última aplicação
	if len(user.Applications()) == 1 {
		return errors.New("user must have access to at least one application")
	}
	
	return nil
}

// ValidateUserUpdate valida se um usuário pode ser atualizado
func (s *userDomainServiceImpl) ValidateUserUpdate(user entities.User, newEmail valueobjects.Email, newUsername valueobjects.Username) error {
	if !user.IsActive() {
		return errors.New("only active users can be updated")
	}
	
	// Verificar se o novo email é válido
	if newEmail.IsEmpty() {
		return errors.New("new email cannot be empty")
	}
	
	// Verificar se o novo username é válido
	if err := s.ValidateUsername(newUsername); err != nil {
		return err
	}
	
	return nil
}
