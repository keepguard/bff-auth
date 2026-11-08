package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/pkg"
)

type resetFake struct{ err error }

func (f resetFake) Execute(context.Context, appdto.SendResetPasswordMessageCommand) (appdto.SendResetPasswordMessageViewDTO, error) {
	if f.err != nil {
		return appdto.SendResetPasswordMessageViewDTO{}, f.err
	}
	return appdto.SendResetPasswordMessageViewDTO{Success: true, Message: appdto.ResetPasswordGenericMessage}, nil
}

func pedirReset(t *testing.T, err error) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", strings.NewReader(`{"email":"alguem@exemplo.com"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "e17ec533-44ea-47ad-af7a-81db7b4e6a8f")
	rec := httptest.NewRecorder()
	h := NewMessageHandlers(resetFake{err: err}, zap.NewNop())
	if e := h.SendResetPasswordMessageHandler(echo.New().NewContext(req, rec)); e != nil {
		t.Fatalf("handler: %v", e)
	}
	return rec
}

// Com conta, sem conta ou com conta inativa, a resposta é a mesma: não dá para descobrir quais e-mails têm conta.
func TestEsqueciSenhaNaoRevelaSeAContaExiste(t *testing.T) {
	sucesso := pedirReset(t, nil)
	casos := map[string]error{
		"sem conta (404)": &appdto.HTTPError{StatusCode: http.StatusNotFound, ErrorCode: "USER_NOT_FOUND", Message: "x"},
		"conta inativa":   &pkg.AppError{StatusCode: http.StatusBadRequest, Code: "USER_NOT_ACTIVE", Message: "x"},
	}
	for nome, err := range casos {
		rec := pedirReset(t, err)
		if rec.Code != sucesso.Code || rec.Body.String() != sucesso.Body.String() {
			t.Errorf("%s: %d %s; esperava igual ao sucesso: %d %s", nome, rec.Code, rec.Body.String(), sucesso.Code, sucesso.Body.String())
		}
	}
	if sucesso.Code != http.StatusOK || !strings.Contains(sucesso.Body.String(), appdto.ResetPasswordGenericMessage) {
		t.Errorf("sucesso: %d %s", sucesso.Code, sucesso.Body.String())
	}
}

// Falha de verdade (serviço fora) continua sendo erro: a resposta genérica é só para "sem conta".
func TestEsqueciSenhaFalhaRealContinuaErro(t *testing.T) {
	rec := pedirReset(t, errors.New("ms-user fora do ar"))
	if rec.Code == http.StatusOK {
		t.Errorf("falha real virou 200: %s", rec.Body.String())
	}
}
