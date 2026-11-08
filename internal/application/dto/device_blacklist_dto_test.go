package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

// O IP do aparelho bloqueado vem do ms-auth e precisa chegar ao front (antes o DTO descartava).
func TestDeviceBlacklistDTOLevaOIP(t *testing.T) {
	doMsAuth := `[{"codeUser":"u1","deviceId":"dev_1","deviceName":"Chrome no macOS","ipAddress":"189.45.12.8",
		"userAgent":"UA","reason":"perdido","blockedAt":"2026-11-08T10:00:00Z","blockedBy":"u1","expiresAt":"2026-12-08T10:00:00Z"}]`
	var lista []DeviceBlacklistDTO
	if err := json.Unmarshal([]byte(doMsAuth), &lista); err != nil {
		t.Fatal(err)
	}
	saida, _ := json.Marshal(lista)
	for _, campo := range []string{`"ipAddress":"189.45.12.8"`, `"userAgent":"UA"`, `"expiresAt":"2026-12-08T10:00:00Z"`} {
		if !strings.Contains(string(saida), campo) {
			t.Errorf("faltou %s em %s", campo, saida)
		}
	}
}
