package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCompanyName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid company name",
			value:   "Test Company",
			wantErr: false,
		},
		{
			name:    "company name with special characters",
			value:   "Company & Associates Ltd.",
			wantErr: false,
		},
		{
			name:    "empty company name",
			value:   "",
			wantErr: true,
			errMsg:  "company name cannot be empty",
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
			errMsg:  "company name cannot be empty",
		},
		{
			name:    "too short",
			value:   "A",
			wantErr: true,
			errMsg:  "company name must be at least 2 characters long",
		},
		{
			name:    "too long",
			value:   generateLongString(101),
			wantErr: true,
			errMsg:  "company name must be at most 100 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companyName, err := NewCompanyName(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, companyName.IsEmpty())
			} else {
				assert.NoError(t, err)
				assert.False(t, companyName.IsEmpty())
				assert.Equal(t, tt.value, companyName.Value())
			}
		})
	}
}

func TestCompanyName_String(t *testing.T) {
	companyName, _ := NewCompanyName("Test Company")
	assert.Equal(t, "Test Company", companyName.String())
}

func TestCompanyName_Value(t *testing.T) {
	companyName, _ := NewCompanyName("Test Company")
	assert.Equal(t, "Test Company", companyName.Value())
}

func TestCompanyName_Equals(t *testing.T) {
	companyName1, _ := NewCompanyName("Test Company")
	companyName2, _ := NewCompanyName("Test Company")
	companyName3, _ := NewCompanyName("Another Company")

	assert.True(t, companyName1.Equals(companyName2))
	assert.False(t, companyName1.Equals(companyName3))
}

func TestCompanyName_IsEmpty(t *testing.T) {
	companyName, _ := NewCompanyName("Test Company")
	assert.False(t, companyName.IsEmpty())

	emptyCompanyName, _ := NewCompanyName("")
	assert.True(t, emptyCompanyName.IsEmpty())
}

