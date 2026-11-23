// Package locale normaliza o idioma da conta para o contrato de mensagens do
// Echouno (pt-BR, en-US, es-ES). A saída é sempre uma constante desta lista:
// o valor cru nunca segue adiante.
package locale

import "strings"

const (
	PtBR = "pt-BR"
	EnUS = "en-US"
	EsES = "es-ES"
)

// Normalize aplica a regra fechada do plano multilíngue:
// pt* → pt-BR, es* → es-ES, qualquer outro idioma → en-US.
// Vazio ou só espaços → "" (sem idioma: quem chama omite o campo).
func Normalize(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	primary := strings.ToLower(raw)
	if i := strings.IndexAny(primary, "-_"); i >= 0 {
		primary = primary[:i]
	}
	switch primary {
	case "pt":
		return PtBR
	case "es":
		return EsES
	default:
		return EnUS
	}
}
