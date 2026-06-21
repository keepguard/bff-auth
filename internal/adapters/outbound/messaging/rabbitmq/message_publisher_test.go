package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNewMessagePublisher(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.RabbitMQConfig
		expectError bool
	}{
		{
			name: "configuração válida",
			config: &config.RabbitMQConfig{
				Host:       "localhost",
				Port:       5672,
				User:       "guest",
				Password:   "guest",
				VHost:      "/",
				Exchange:   "test-exchange",
				RoutingKey: "test.routing.key",
				Durable:    true,
				AutoDelete: false,
			},
			expectError: false,
		},
		{
			name:        "configuração nula",
			config:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			publisher, err := NewMessagePublisher(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, publisher)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, publisher)
			}
		})
	}
}

func TestMessagePublisher_MessageSerialization(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := &config.RabbitMQConfig{
		Host:       "localhost",
		Port:       5672,
		User:       "guest",
		Password:   "guest",
		VHost:      "/",
		Exchange:   "test-exchange",
		RoutingKey: "test.routing.key",
		Durable:    true,
		AutoDelete: false,
	}

	publisher, err := NewMessagePublisher(cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, publisher)

	// Testar serialização de diferentes tipos de variáveis
	message := messaging.MessageDTO{
		XApplication:      "test-app",
		XCorrelationID:    "test-correlation-id",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RESET_PASSWORD",
		Recipient:         "test@example.com",
		CodeUser:          "user123",
		Subject:           "Test Subject",
		Content:           "Test Content",
		Variables: map[string]interface{}{
			"string": "test string",
			"int":    42,
			"float":  3.14,
			"bool":   true,
			"array":  []string{"item1", "item2"},
			"object": map[string]string{"key": "value"},
		},
	}

	// Verificar se a mensagem pode ser serializada
	messageBytes, err := json.Marshal(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, messageBytes)

	// Verificar se pode ser deserializada
	var deserializedMessage messaging.MessageDTO
	err = json.Unmarshal(messageBytes, &deserializedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message.XApplication, deserializedMessage.XApplication)
	assert.Equal(t, message.XCorrelationID, deserializedMessage.XCorrelationID)
	assert.Equal(t, message.MessageType, deserializedMessage.MessageType)
	assert.Equal(t, message.CommunicationType, deserializedMessage.CommunicationType)
	assert.Equal(t, message.TemplateType, deserializedMessage.TemplateType)
	assert.Equal(t, message.Recipient, deserializedMessage.Recipient)
	assert.Equal(t, message.CodeUser, deserializedMessage.CodeUser)
	assert.Equal(t, message.Subject, deserializedMessage.Subject)
	assert.Equal(t, message.Content, deserializedMessage.Content)
	assert.Equal(t, message.Variables["string"], deserializedMessage.Variables["string"])
	assert.Equal(t, message.Variables["bool"], deserializedMessage.Variables["bool"])
	assert.Equal(t, message.Variables["float"], deserializedMessage.Variables["float"])
	// Nota: JSON unmarshal converte números para float64
	assert.Equal(t, float64(42), deserializedMessage.Variables["int"])
	// Nota: JSON unmarshal converte arrays para []interface{}
	assert.Equal(t, []interface{}{"item1", "item2"}, deserializedMessage.Variables["array"])
	// Nota: JSON unmarshal converte maps para map[string]interface{}
	assert.Equal(t, map[string]interface{}{"key": "value"}, deserializedMessage.Variables["object"])
}

func TestMessagePublisher_SerializationError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := &config.RabbitMQConfig{
		Host:       "localhost",
		Port:       5672,
		User:       "guest",
		Password:   "guest",
		VHost:      "/",
		Exchange:   "test-exchange",
		RoutingKey: "test.routing.key",
		Durable:    true,
		AutoDelete: false,
	}

	publisher, err := NewMessagePublisher(cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, publisher)

	// Criar mensagem com dados que causam erro de serialização
	message := messaging.MessageDTO{
		XApplication:      "test-app",
		XCorrelationID:    "test-correlation-id",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RESET_PASSWORD",
		Recipient:         "test@example.com",
		CodeUser:          "user123",
		Variables: map[string]interface{}{
			"invalid": make(chan int), // Canal não pode ser serializado para JSON
		},
	}

	err = publisher.PublishMessage(context.Background(), message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao serializar mensagem")
}
