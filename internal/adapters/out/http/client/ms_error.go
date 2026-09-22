package client

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/pkg"
)

func buildHTTPErrorFromResponse(resp *resty.Response, operation string) error {
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return nil
	}

	var msError appdto.MSAuthErrorResponse
	if err := json.Unmarshal(resp.Body(), &msError); err == nil && (msError.Title != "" || msError.Detail != "") {
		rawMessage := msError.Detail
		if rawMessage == "" {
			rawMessage = msError.Title
		}
		errorCode := getErrorCodeFromProperties(msError.Properties)
		message := pkg.ResolveUserMessage(errorCode, rawMessage)
		if message == "" {
			message = rawMessage
		}
		statusCode := msError.Status
		if statusCode == 0 {
			statusCode = resp.StatusCode()
		}
		return &appdto.HTTPError{
			StatusCode: statusCode,
			Message:    message,
			ErrorCode:  errorCode,
			Details:    msError.Properties,
		}
	}

	return &appdto.HTTPError{
		StatusCode: resp.StatusCode(),
		Message:    fmt.Sprintf("Erro ao comunicar com o serviço de autenticação (operação: %s)", operation),
		ErrorCode:  "HTTP_ERROR",
		Details:    map[string]interface{}{"body": string(resp.Body())},
	}
}
