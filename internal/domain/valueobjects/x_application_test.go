package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewXApplication(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid x-application",
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
			name:    "empty x-application",
			value:   "",
			wantErr: true,
			errMsg:  "x-application cannot be empty",
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
			errMsg:  "x-application cannot be empty",
		},
		{
			name:    "too short",
			value:   "ab",
			wantErr: true,
			errMsg:  "x-application must be at least 3 characters long",
		},
		{
			name:    "too long",
			value:   generateLongString(51),
			wantErr: true,
			errMsg:  "x-application must be at most 50 characters long",
		},
		{
			name:    "invalid characters",
			value:   "my@app",
			wantErr: true,
			errMsg:  "x-application can only contain letters, numbers, underscore and hyphen",
		},
		{
			name:    "invalid characters with spaces",
			value:   "my app",
			wantErr: true,
			errMsg:  "x-application can only contain letters, numbers, underscore and hyphen",
		},
		{
			name:    "invalid characters with special chars",
			value:   "my.app",
			wantErr: true,
			errMsg:  "x-application can only contain letters, numbers, underscore and hyphen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xApplication, err := NewXApplication(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, xApplication.IsEmpty())
			} else {
				assert.NoError(t, err)
				assert.False(t, xApplication.IsEmpty())
				// X-Application deve ser convertido para lowercase
				assert.Equal(t, strings.ToLower(tt.value), xApplication.Value())
			}
		})
	}
}

func TestXApplication_String(t *testing.T) {
	xApplication, _ := NewXApplication("MyApp")
	assert.Equal(t, "myapp", xApplication.String()) // Deve ser lowercase
}

func TestXApplication_Value(t *testing.T) {
	xApplication, _ := NewXApplication("MyApp")
	assert.Equal(t, "myapp", xApplication.Value()) // Deve ser lowercase
}

func TestXApplication_Equals(t *testing.T) {
	xApplication1, _ := NewXApplication("MyApp")
	xApplication2, _ := NewXApplication("myapp")
	xApplication3, _ := NewXApplication("AnotherApp")

	assert.True(t, xApplication1.Equals(xApplication2)) // Case insensitive
	assert.False(t, xApplication1.Equals(xApplication3))
}

func TestXApplication_IsEmpty(t *testing.T) {
	xApplication, _ := NewXApplication("myapp")
	assert.False(t, xApplication.IsEmpty())

	emptyXApplication, _ := NewXApplication("")
	assert.True(t, emptyXApplication.IsEmpty())
}
