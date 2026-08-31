package pkg

import "testing"

func TestHasAnyRole(t *testing.T) {
	if !HasAnyRole([]string{"ROLE_ADMIN"}, "ADMIN", "SYSTEM") {
		t.Fatal("expected ROLE_ADMIN to match ADMIN")
	}
	if !HasAnyRole([]string{"MANAGER"}, "ADMIN", "MANAGER") {
		t.Fatal("expected MANAGER to match")
	}
	if HasAnyRole([]string{"ROLE_USER"}, "ADMIN", "SYSTEM", "MANAGER") {
		t.Fatal("USER must not match privileged roles")
	}
}
