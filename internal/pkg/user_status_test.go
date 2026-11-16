package pkg

import "testing"

// Tabela por status: só ACTIVE passa. PENDING_ACTIVATION (cadastro em andamento) é recusado.
func TestIsActiveUserStatus(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{status: "ACTIVE", want: true},
		{status: "PENDING_ACTIVATION", want: false},
		{status: "PENDING", want: false},
		{status: "BLOCKED", want: false},
		{status: "DELETED", want: false},
		{status: "INACTIVE", want: false},
		{status: "active", want: false}, // o ms-user responde em maiúsculas; minúsculas não são ACTIVE
		{status: " ACTIVE", want: false},
		{status: "", want: false},
	}
	for _, tc := range cases {
		t.Run("status="+tc.status, func(t *testing.T) {
			if got := IsActiveUserStatus(tc.status); got != tc.want {
				t.Fatalf("IsActiveUserStatus(%q) = %v, quer %v", tc.status, got, tc.want)
			}
		})
	}
}
