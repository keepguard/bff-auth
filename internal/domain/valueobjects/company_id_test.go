package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCompanyID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid company ID",
			value:   "company-123",
			wantErr: false,
		},
		{
			name:    "valid UUID format",
			value:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "empty company ID",
			value:   "",
			wantErr: true,
			errMsg:  "company ID cannot be empty",
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
			errMsg:  "company ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companyID, err := NewCompanyID(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, companyID.IsEmpty())
			} else {
				assert.NoError(t, err)
				assert.False(t, companyID.IsEmpty())
				assert.Equal(t, tt.value, companyID.Value())
			}
		})
	}
}

func TestCompanyID_String(t *testing.T) {
	companyID, _ := NewCompanyID("company-123")
	assert.Equal(t, "company-123", companyID.String())
}

func TestCompanyID_Value(t *testing.T) {
	companyID, _ := NewCompanyID("company-123")
	assert.Equal(t, "company-123", companyID.Value())
}

func TestCompanyID_Equals(t *testing.T) {
	companyID1, _ := NewCompanyID("company-123")
	companyID2, _ := NewCompanyID("company-123")
	companyID3, _ := NewCompanyID("company-456")

	assert.True(t, companyID1.Equals(companyID2))
	assert.False(t, companyID1.Equals(companyID3))
}

func TestCompanyID_IsEmpty(t *testing.T) {
	companyID, _ := NewCompanyID("company-123")
	assert.False(t, companyID.IsEmpty())

	emptyCompanyID, _ := NewCompanyID("")
	assert.True(t, emptyCompanyID.IsEmpty())
}
