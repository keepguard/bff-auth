package pkg

import "strings"

func NormalizeRole(role string) string {
	normalized := strings.ToUpper(strings.TrimSpace(role))
	return strings.TrimPrefix(normalized, "ROLE_")
}

func HasAnyRole(have []string, want ...string) bool {
	if len(have) == 0 || len(want) == 0 {
		return false
	}
	wanted := make(map[string]struct{}, len(want))
	for _, role := range want {
		if n := NormalizeRole(role); n != "" {
			wanted[n] = struct{}{}
		}
	}
	for _, role := range have {
		if _, ok := wanted[NormalizeRole(role)]; ok {
			return true
		}
	}
	return false
}
