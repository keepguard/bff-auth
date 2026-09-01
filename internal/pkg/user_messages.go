package pkg

import "strings"

var errorCodeMessages = map[string]string{
	"USER_NOT_FOUND":             "Usuário não encontrado",
	"INVALID_PASSWORD":           "Senha inválida",
	"INVALID_CREDENTIALS":        "Credenciais inválidas",
	"USER_NOT_ACTIVE":            "Usuário não está ativo",
	"EMAIL_NOT_VERIFIED":         "E-mail não verificado",
	"INVALID_TOKEN":              "Token inválido ou expirado",
	"TOKEN_REVOKED":              "Sessão revogada ou expirada. Por favor, realize login novamente.",
	"SESSION_ALREADY_TERMINATED": "Sessão já encerrada",
	"INVALID_EMAIL":              "E-mail inválido",
	"INVALID_PASSWORD_FORMAT":    "Senha inválida",
	"INVALID_TENANT_ID":          "Header X-Tenant-Id inválido",
	"DEVICE_BLACKLISTED":         "Este dispositivo foi bloqueado para acesso a esta conta.",
	"RESOURCE_NOT_FOUND":         "Recurso não encontrado",
	"ALREADY_EXISTS":             "Recurso já existe",
	"REQUIRED_FIELD":             "Campo obrigatório ausente",
	"COMPANY_NOT_FOUND":          "Empresa não encontrada",
	"FORBIDDEN":                  "Acesso negado",
	"SESSION_FORBIDDEN":          "Operação não permitida para o seu perfil",
	"SELF_ONLY":                  "Você só pode consultar a própria conta",
	"TENANT_MISMATCH":            "Usuário pertence a outro tenant",
	"CHALLENGE_NOT_FOUND":        "Sessão de desafio expirada ou inválida",
	"INVALID_CHALLENGE_CODE":     "Código de segurança inválido",
	"MAX_ATTEMPTS_EXCEEDED":      "Número máximo de tentativas excedido. Tente fazer login novamente.",
	"INVALID_REVOKE_TOKEN":       "Link de revocação inválido ou expirado",
	"ACCOUNT_LOCKED":             "Conta temporariamente bloqueada",
}

var englishDetailMessages = map[string]string{
	"User not found":                     "Usuário não encontrado",
	"Invalid password":                   "Senha inválida",
	"User is not active":                 "Usuário não está ativo",
	"Email not verified":                 "E-mail não verificado",
	"Invalid token":                      "Token inválido ou expirado",
	"Invalid or expired reset token":     "Token de reset inválido ou expirado",
	"New password and confirmation do not match": "A nova senha e a confirmação não coincidem",
	"Current password is incorrect":      "Senha atual incorreta",
	"New password cannot be the same as one of the last 5 passwords": "A nova senha não pode ser igual a uma das últimas 5 senhas",
	"Bad Request": "Dados de entrada inválidos",
}

// ResolveUserMessage devolve a mensagem em português para exibição ao usuário.
// Quando rawMessage está presente, prioriza o texto (traduzindo se estiver em inglês).
// Caso contrário, usa errorCode como fallback.
func ResolveUserMessage(errorCode, rawMessage string) string {
	raw := strings.TrimSpace(rawMessage)
	if raw != "" {
		if msg, ok := englishDetailMessages[raw]; ok {
			return msg
		}
		return raw
	}

	code := strings.ToUpper(strings.TrimSpace(errorCode))
	if code != "" && code != "UNKNOWN_ERROR" && code != "HTTP_ERROR" {
		if msg, ok := errorCodeMessages[code]; ok {
			return msg
		}
	}
	return ""
}
