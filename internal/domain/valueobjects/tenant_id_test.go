package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTenantId(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid x-tenant-id",
			value:   "my-app",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			value:   "my_app",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			value:   "app123",
			wantErr: false,
		},
		{
			name:    "valid with uppercase",
			value:   "MyApp",
			wantErr: false,
		},
		{
			name:    "empty x-tenant-id",
			value:   "",
			wantErr: true,
			errMsg:  "x-tenant-id cannot be empty",
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
			errMsg:  "x-tenant-id cannot be empty",
		},
		{
			name:    "too short",
			value:   "ab",
			wantErr: true,
			errMsg:  "x-tenant-id must be at least 3 characters long",
		},
		{
			name:    "too long",
			value:   generateLongString(51),
			wantErr: true,
			errMsg:  "x-tenant-id must be at most 50 characters long",
		},
		{
			name:    "invalid characters",
			value:   "my@app",
			wantErr: true,
			errMsg:  "x-tenant-id can only contain letters, numbers, underscore and hyphen",
		},
		{
			name:    "invalid characters with spaces",
			value:   "my app",
			wantErr: true,
			errMsg:  "x-tenant-id can only contain letters, numbers, underscore and hyphen",
		},
		{
			name:    "invalid characters with special chars",
			value:   "my.app",
			wantErr: true,
			errMsg:  "x-tenant-id can only contain letters, numbers, underscore and hyphen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantId, err := NewTenantId(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, tenantId.IsEmpty())
			} else {
				assert.NoError(t, err)
				assert.False(t, tenantId.IsEmpty())
				// X-Tenant-Id deve ser convertido para lowercase
				assert.Equal(t, strings.ToLower(tt.value), tenantId.Value())
			}
		})
	}
}

func TestTenantId_String(t *testing.T) {
	tenantId, _ := NewTenantId("MyApp")
	assert.Equal(t, "myapp", tenantId.String()) // Deve ser lowercase
}

func TestTenantId_Value(t *testing.T) {
	tenantId, _ := NewTenantId("MyApp")
	assert.Equal(t, "myapp", tenantId.Value()) // Deve ser lowercase
}

func TestTenantId_Equals(t *testing.T) {
	tenantId1, _ := NewTenantId("MyApp")
	tenantId2, _ := NewTenantId("myapp")
	tenantId3, _ := NewTenantId("AnotherApp")

	assert.True(t, tenantId1.Equals(tenantId2)) // Case insensitive
	assert.False(t, tenantId1.Equals(tenantId3))
}

func TestTenantId_IsEmpty(t *testing.T) {
	tenantId, _ := NewTenantId("myapp")
	assert.False(t, tenantId.IsEmpty())

	emptyTenantId, _ := NewTenantId("")
	assert.True(t, emptyTenantId.IsEmpty())
}
