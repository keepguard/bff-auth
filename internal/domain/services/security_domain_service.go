package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

// SecurityDomainService define as regras de negócio para segurança
type SecurityDomainService interface {
	ValidatePasswordStrength(password string) error
	GenerateSecureToken() string
	ValidatePasswordHistory(password string, history []string) error
	CalculatePasswordEntropy(password string) float64
	ShouldRequirePasswordChange(lastChange time.Time) bool
	ValidatePasswordComplexity(password string) error
	GenerateSecureRandomString(length int) string
	ValidateSecurityHeaders(headers map[string]string) error
	ShouldBlockUser(attempts int, lastAttempt time.Time) bool
	CalculatePasswordScore(password string) int
}

type securityDomainServiceImpl struct{}

// NewSecurityDomainService cria um novo serviço de domínio para segurança
func NewSecurityDomainService() SecurityDomainService {
	return &securityDomainServiceImpl{}
}

// ValidatePasswordStrength valida a força de uma senha
func (s *securityDomainServiceImpl) ValidatePasswordStrength(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	
	// Verificar comprimento mínimo
	if len(password) < 8 {
		return errors.New("password must have at least 8 characters")
	}
	
	// Verificar comprimento máximo
	if len(password) > 128 {
		return errors.New("password must have at most 128 characters")
	}
	
	// Verificar complexidade
	if err := s.ValidatePasswordComplexity(password); err != nil {
		return err
	}
	
	// Verificar se não contém sequências comuns
	if s.containsCommonSequences(password) {
		return errors.New("password contains common sequences")
	}
	
	// Verificar se não contém informações pessoais comuns
	if s.containsPersonalInfo(password) {
		return errors.New("password contains personal information")
	}
	
	return nil
}

// GenerateSecureToken gera um token seguro
func (s *securityDomainServiceImpl) GenerateSecureToken() string {
	// Gera 32 bytes aleatórios
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidatePasswordHistory valida se a senha não foi usada recentemente
func (s *securityDomainServiceImpl) ValidatePasswordHistory(password string, history []string) error {
	if len(history) == 0 {
		return nil // Sem histórico, pode usar qualquer senha
	}
	
	// Verificar se a senha não está no histórico (últimas 5 senhas)
	maxHistory := 5
	if len(history) > maxHistory {
		history = history[:maxHistory]
	}
	
	for _, oldPassword := range history {
		if password == oldPassword {
			return errors.New("password was used recently")
		}
	}
	
	return nil
}

// CalculatePasswordEntropy calcula a entropia de uma senha
func (s *securityDomainServiceImpl) CalculatePasswordEntropy(password string) float64 {
	if password == "" {
		return 0
	}
	
	// Contar tipos de caracteres
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
	
	// Calcular tamanho do alfabeto
	alphabetSize := 0
	if hasLower {
		alphabetSize += 26
	}
	if hasUpper {
		alphabetSize += 26
	}
	if hasNumber {
		alphabetSize += 10
	}
	if hasSpecial {
		alphabetSize += 32 // Aproximadamente 32 caracteres especiais
	}
	
	// Calcular entropia: log2(alphabetSize^length)
	if alphabetSize == 0 {
		return 0
	}
	
	entropy := float64(len(password)) * (1.0 / 0.6931471805599453) * float64(alphabetSize) // log2
	return entropy
}

// ShouldRequirePasswordChange verifica se deve exigir mudança de senha
func (s *securityDomainServiceImpl) ShouldRequirePasswordChange(lastChange time.Time) bool {
	// Exigir mudança se a senha tem mais de 90 dias
	return time.Since(lastChange) > 90*24*time.Hour
}

// ValidatePasswordComplexity valida a complexidade de uma senha
func (s *securityDomainServiceImpl) ValidatePasswordComplexity(password string) error {
	// Verificar se contém pelo menos uma letra minúscula
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	
	// Verificar se contém pelo menos uma letra maiúscula
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	
	// Verificar se contém pelo menos um número
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one number")
	}
	
	// Verificar se contém pelo menos um caractere especial
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}
	
	return nil
}

// GenerateSecureRandomString gera uma string aleatória segura
func (s *securityDomainServiceImpl) GenerateSecureRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)
	
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[bytes[i]%byte(len(charset))]
	}
	
	return string(result)
}

// ValidateSecurityHeaders valida headers de segurança
func (s *securityDomainServiceImpl) ValidateSecurityHeaders(headers map[string]string) error {
	requiredHeaders := []string{
		"X-Application",
		"X-Correlation-ID",
	}
	
	for _, header := range requiredHeaders {
		if headers[header] == "" {
			return errors.New("missing required security header: " + header)
		}
	}
	
	// Validar X-Application
	if err := s.validateXApplicationHeader(headers["X-Application"]); err != nil {
		return err
	}
	
	// Validar X-Correlation-ID
	if err := s.validateCorrelationIDHeader(headers["X-Correlation-ID"]); err != nil {
		return err
	}
	
	return nil
}

// ShouldBlockUser verifica se um usuário deve ser bloqueado
func (s *securityDomainServiceImpl) ShouldBlockUser(attempts int, lastAttempt time.Time) bool {
	// Bloquear se mais de 5 tentativas em 15 minutos
	if attempts >= 5 && time.Since(lastAttempt) < 15*time.Minute {
		return true
	}
	
	// Bloquear se mais de 10 tentativas em 1 hora
	if attempts >= 10 && time.Since(lastAttempt) < 1*time.Hour {
		return true
	}
	
	return false
}

// CalculatePasswordScore calcula um score de força da senha (0-100)
func (s *securityDomainServiceImpl) CalculatePasswordScore(password string) int {
	if password == "" {
		return 0
	}
	
	score := 0
	
	// Pontuação por comprimento
	length := len(password)
	if length >= 8 {
		score += 20
	}
	if length >= 12 {
		score += 10
	}
	if length >= 16 {
		score += 10
	}
	
	// Pontuação por tipos de caracteres
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		score += 10
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score += 10
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		score += 10
	}
	if regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		score += 10
	}
	
	// Pontuação por complexidade
	entropy := s.CalculatePasswordEntropy(password)
	if entropy > 50 {
		score += 10
	}
	if entropy > 70 {
		score += 10
	}
	
	// Penalizar sequências comuns
	if s.containsCommonSequences(password) {
		score -= 20
	}
	
	// Garantir que o score esteja entre 0 e 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

// containsCommonSequences verifica se a senha contém sequências comuns
func (s *securityDomainServiceImpl) containsCommonSequences(password string) bool {
	commonSequences := []string{
		"123", "abc", "qwe", "asd", "zxc",
		"password", "123456", "qwerty", "admin",
		"letmein", "welcome", "monkey", "dragon",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, sequence := range commonSequences {
		if strings.Contains(lowerPassword, sequence) {
			return true
		}
	}
	
	return false
}

// containsPersonalInfo verifica se a senha contém informações pessoais
func (s *securityDomainServiceImpl) containsPersonalInfo(password string) bool {
	// Implementação simplificada - em produção, verificar contra dados do usuário
	personalInfo := []string{
		"username", "email", "name", "birthday",
		"company", "phone", "address",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, info := range personalInfo {
		if strings.Contains(lowerPassword, info) {
			return true
		}
	}
	
	return false
}

// validateXApplicationHeader valida o header X-Application
func (s *securityDomainServiceImpl) validateXApplicationHeader(xApplication string) error {
	if len(xApplication) < 3 || len(xApplication) > 50 {
		return errors.New("invalid X-Application header length")
	}
	
	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(xApplication) {
		return errors.New("invalid X-Application header format")
	}
	
	return nil
}

// validateCorrelationIDHeader valida o header X-Correlation-ID
func (s *securityDomainServiceImpl) validateCorrelationIDHeader(correlationID string) error {
	if len(correlationID) < 10 || len(correlationID) > 100 {
		return errors.New("invalid X-Correlation-ID header length")
	}
	
	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(correlationID) {
		return errors.New("invalid X-Correlation-ID header format")
	}
	
	return nil
}
