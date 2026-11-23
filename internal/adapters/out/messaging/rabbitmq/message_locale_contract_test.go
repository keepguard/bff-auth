package rabbitmq

import (
	"encoding/json"
	"testing"

	outport "github.com/keepguard/bff-auth/internal/application/port/out"
	"github.com/stretchr/testify/assert"
)

func resetMessage() outport.MessageDTO {
	return outport.MessageDTO{
		CompanyID:         "company-123",
		CorrelationID:     "corr-1",
		XCorrelationID:    "corr-1",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Recipient:         "a@b.com",
		CodeUser:          "code-user-1",
		Variables:         map[string]interface{}{"userName": "x", "token": "123456"},
	}
}

// Sem idioma, o JSON da fila é byte a byte o de antes da mudança.
func TestMessageDTO_WithoutLocale_JSONIdenticalToPreviousContract(t *testing.T) {
	raw, err := json.Marshal(resetMessage())
	assert.NoError(t, err)
	assert.Equal(t,
		`{"companyId":"company-123","correlationId":"corr-1","xCorrelationId":"corr-1","messageType":"EMAIL","communicationType":"EMAIL","templateType":"RECUPERACAO_SENHA","recipient":"a@b.com","codeUser":"code-user-1","variables":{"token":"123456","userName":"x"}}`,
		string(raw))
}

// Com idioma, o campo aparece no topo do JSON, com o nome que o ms-communication lê.
func TestMessageDTO_WithLocale_EmitsTopLevelLocale(t *testing.T) {
	msg := resetMessage()
	msg.Locale = "en-US"
	raw, err := json.Marshal(msg)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	assert.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, "en-US", decoded["locale"])
}
