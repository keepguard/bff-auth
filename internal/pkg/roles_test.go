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

func TestHasAuthority(t *testing.T) {
	if !HasAuthority([]string{"session:read", "collector:read"}, "session:read") {
		t.Fatal("expected session:read to match")
	}
	if HasAuthority([]string{"session:read"}, "session:write") {
		t.Fatal("read must not match write")
	}
	if HasAuthority(nil, "session:read") {
		t.Fatal("empty authorities must not match")
	}
}
