package services

import (
	"testing"

	"github.com/keepguard/bff-auth/internal/domain/entities"
	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestNewUserDomainService(t *testing.T) {
	service := NewUserDomainService()
	assert.NotNil(t, service)
	assert.IsType(t, &userDomainServiceImpl{}, service)
}

func TestUserDomainService_ValidateUserCreation(t *testing.T) {
	service := NewUserDomainService()

	tests := []struct {
		name    string
		user    entities.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user creation",
			user: func() entities.User {
				userID, _ := valueobjects.NewUserID("user-123")
				username, _ := valueobjects.NewUsername("testuser")
				email, _ := valueobjects.NewEmail("test@example.com")
				return *entities.NewUser(userID, username, email)
			}(),
			wantErr: false,
		},
		{
			name: "empty user ID",
			user: func() entities.User {
				userID, _ := valueobjects.NewUserID("")
				username, _ := valueobjects.NewUsername("testuser")
				email, _ := valueobjects.NewEmail("test@example.com")
				return *entities.NewUser(userID, username, email)
			}(),
			wantErr: true,
			errMsg:  "user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateUserCreation(tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserDomainService_ValidatePassword(t *testing.T) {
	service := NewUserDomainService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password",
			password: "MySecure123!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
			errMsg:   "password is required",
		},
		{
			name:     "password too short",
			password: "My123!",
			wantErr:  true,
			errMsg:   "password must have at least 8 characters",
		},
		{
			name:     "password too long",
			password: "MySecure123!ThisIsAVeryLongPasswordThatExceedsTheMaximumLengthAllowedByTheSystemAndShouldBeRejectedBecauseItIsTooLongAndViolatesThePasswordPolicy",
			wantErr:  true,
			errMsg:   "password must have at most 128 characters",
		},
		{
			name:     "password without lowercase",
			password: "MYSECURE123!",
			wantErr:  true,
			errMsg:   "password must contain at least one lowercase letter",
		},
		{
			name:     "password without uppercase",
			password: "mysecure123!",
			wantErr:  true,
			errMsg:   "password must contain at least one uppercase letter",
		},
		{
			name:     "password without number",
			password: "MySecurePassword!",
			wantErr:  true,
			errMsg:   "password must contain at least one number",
		},
		{
			name:     "password without special character",
			password: "MySecure123",
			wantErr:  true,
			errMsg:   "password must contain at least one special character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserDomainService_CanUserBeActivated(t *testing.T) {
	service := NewUserDomainService()

	userID, _ := valueobjects.NewUserID("user-123")
	username, _ := valueobjects.NewUsername("testuser")
	email, _ := valueobjects.NewEmail("test@example.com")

	user := entities.NewUser(userID, username, email)
	user.Deactivate() // Garante que está inativo

	err := service.CanUserBeActivated(*user)
	assert.NoError(t, err)

	// Teste com usuário já ativo
	user.Activate()
	err = service.CanUserBeActivated(*user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user is already active")
}

func TestUserDomainService_CanUserBeDeactivated(t *testing.T) {
	service := NewUserDomainService()

	userID, _ := valueobjects.NewUserID("user-123")
	username, _ := valueobjects.NewUsername("testuser")
	email, _ := valueobjects.NewEmail("test@example.com")

	user := entities.NewUser(userID, username, email)
	user.Activate() // Garante que está ativo

	err := service.CanUserBeDeactivated(*user)
	assert.NoError(t, err)

	// Teste com usuário já inativo
	user.Deactivate()
	err = service.CanUserBeDeactivated(*user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user is already inactive")
}

func TestUserDomainService_GenerateUsername(t *testing.T) {
	service := NewUserDomainService()

	email, _ := valueobjects.NewEmail("joao.silva@example.com")
	username := service.GenerateUsername(email)

	assert.False(t, username.IsEmpty())
	assert.True(t, len(username.Value()) >= 3)
	assert.True(t, len(username.Value()) <= 50)
}

func TestUserDomainService_ValidateUsername(t *testing.T) {
	service := NewUserDomainService()

	tests := []struct {
		name     string
		username valueobjects.Username
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid username",
			username: func() valueobjects.Username { u, _ := valueobjects.NewUsername("testuser"); return u }(),
			wantErr:  false,
		},
		{
			name:     "username with invalid characters",
			username: func() valueobjects.Username { u, _ := valueobjects.NewUsername("test@user"); return u }(),
			wantErr:  true,
			errMsg:   "username can only contain letters, numbers, underscore and hyphen",
		},
		{
			name:     "username starting with underscore",
			username: func() valueobjects.Username { u, _ := valueobjects.NewUsername("_testuser"); return u }(),
			wantErr:  true,
			errMsg:   "username cannot start with underscore or hyphen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateUsername(tt.username)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserDomainService_CanUserAddRole(t *testing.T) {
	service := NewUserDomainService()

	userID, _ := valueobjects.NewUserID("user-123")
	username, _ := valueobjects.NewUsername("testuser")
	email, _ := valueobjects.NewEmail("test@example.com")

	user := entities.NewUser(userID, username, email)
	user.Activate()

	role, _ := valueobjects.NewRole("user")

	err := service.CanUserAddRole(*user, role)
	assert.NoError(t, err)

	// Adicionar o role
	user.AddRole(role)

	// Tentar adicionar o mesmo role novamente
	err = service.CanUserAddRole(*user, role)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already has this role")
}

func TestUserDomainService_CanUserRemoveRole(t *testing.T) {
	service := NewUserDomainService()

	userID, _ := valueobjects.NewUserID("user-123")
	username, _ := valueobjects.NewUsername("testuser")
	email, _ := valueobjects.NewEmail("test@example.com")

	user := entities.NewUser(userID, username, email)
	user.Activate()

	role, _ := valueobjects.NewRole("user")
	user.AddRole(role)

	err := service.CanUserRemoveRole(*user, role)
	assert.NoError(t, err)

	// Remover o role
	user.RemoveRole(role)

	// Tentar remover o mesmo role novamente
	err = service.CanUserRemoveRole(*user, role)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user does not have this role")
}

func TestUserDomainService_CanUserAddApplication(t *testing.T) {
	service := NewUserDomainService()

	userID, _ := valueobjects.NewUserID("user-123")
	username, _ := valueobjects.NewUsername("testuser")
	email, _ := valueobjects.NewEmail("test@example.com")

	user := entities.NewUser(userID, username, email)
	user.Activate()

	app, _ := valueobjects.NewApplication("web")

	err := service.CanUserAddApplication(*user, app)
	assert.NoError(t, err)

	// Adicionar a aplicação
	user.AddApplication(app)

	// Tentar adicionar a mesma aplicação novamente
	err = service.CanUserAddApplication(*user, app)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already has access to this application")
}

func TestUserDomainService_ValidateUserUpdate(t *testing.T) {
	service := NewUserDomainService()

	userID, _ := valueobjects.NewUserID("user-123")
	username, _ := valueobjects.NewUsername("testuser")
	email, _ := valueobjects.NewEmail("test@example.com")

	user := entities.NewUser(userID, username, email)
	user.Activate()

	newEmail, _ := valueobjects.NewEmail("newemail@example.com")
	newUsername, _ := valueobjects.NewUsername("newusername")

	err := service.ValidateUserUpdate(*user, newEmail, newUsername)
	assert.NoError(t, err)

	// Teste com usuário inativo
	user.Deactivate()
	err = service.ValidateUserUpdate(*user, newEmail, newUsername)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only active users can be updated")
}
