package logger

import (
	"go.uber.org/zap"
)

// StandardFields define os campos padrão para logging
type StandardFields struct {
	TraceID       string
	SpanID        string
	RequestID     string
	UserID        string
	Service       string
	Component     string
	Environment   string
	Version       string
	CorrelationID string
}

// ToZapFields converte StandardFields para zap.Fields
func (sf StandardFields) ToZapFields() []zap.Field {
	fields := []zap.Field{
		zap.String("service", sf.Service),
		zap.String("component", sf.Component),
		zap.String("environment", sf.Environment),
		zap.String("version", sf.Version),
	}

	if sf.TraceID != "" {
		fields = append(fields, zap.String("traceId", sf.TraceID))
	}

	if sf.SpanID != "" {
		fields = append(fields, zap.String("spanId", sf.SpanID))
	}

	if sf.RequestID != "" {
		fields = append(fields, zap.String("requestId", sf.RequestID))
	}

	if sf.UserID != "" {
		fields = append(fields, zap.String("userId", sf.UserID))
	}

	if sf.CorrelationID != "" {
		fields = append(fields, zap.String("correlationId", sf.CorrelationID))
	}

	return fields
}

// NewStandardFields cria um StandardFields com valores padrão
func NewStandardFields(service, component, environment, version string) StandardFields {
	return StandardFields{
		Service:     service,
		Component:   component,
		Environment: environment,
		Version:     version,
	}
}

// WithTraceID adiciona traceId ao StandardFields
func (sf StandardFields) WithTraceID(traceID string) StandardFields {
	sf.TraceID = traceID
	return sf
}

// WithSpanID adiciona spanId ao StandardFields
func (sf StandardFields) WithSpanID(spanID string) StandardFields {
	sf.SpanID = spanID
	return sf
}

// WithRequestID adiciona requestId ao StandardFields
func (sf StandardFields) WithRequestID(requestID string) StandardFields {
	sf.RequestID = requestID
	return sf
}

// WithUserID adiciona userId ao StandardFields
func (sf StandardFields) WithUserID(userID string) StandardFields {
	sf.UserID = userID
	return sf
}

// WithCorrelationID adiciona correlationId ao StandardFields
func (sf StandardFields) WithCorrelationID(correlationID string) StandardFields {
	sf.CorrelationID = correlationID
	return sf
}

// HTTPRequestFields define campos específicos para requisições HTTP
type HTTPRequestFields struct {
	StandardFields
	Method       string
	Path         string
	URI          string
	Status       int
	LatencyMs    int64
	Duration     string
	UserAgent    string
	IP           string
	ResponseSize int64
}

// ToZapFields converte HTTPRequestFields para zap.Fields
func (hrf HTTPRequestFields) ToZapFields() []zap.Field {
	fields := hrf.StandardFields.ToZapFields()

	fields = append(fields,
		zap.String("method", hrf.Method),
		zap.String("path", hrf.Path),
		zap.String("uri", hrf.URI),
		zap.Int("status", hrf.Status),
		zap.Int64("latencyMs", hrf.LatencyMs),
		zap.String("duration", hrf.Duration),
		zap.String("userAgent", hrf.UserAgent),
		zap.String("ip", hrf.IP),
		zap.Int64("responseSize", hrf.ResponseSize),
	)

	return fields
}

// NewHTTPRequestFields cria um HTTPRequestFields
func NewHTTPRequestFields(standardFields StandardFields) HTTPRequestFields {
	return HTTPRequestFields{
		StandardFields: standardFields,
	}
}

// WithMethod adiciona method ao HTTPRequestFields
func (hrf HTTPRequestFields) WithMethod(method string) HTTPRequestFields {
	hrf.Method = method
	return hrf
}

// WithPath adiciona path ao HTTPRequestFields
func (hrf HTTPRequestFields) WithPath(path string) HTTPRequestFields {
	hrf.Path = path
	return hrf
}

// WithURI adiciona uri ao HTTPRequestFields
func (hrf HTTPRequestFields) WithURI(uri string) HTTPRequestFields {
	hrf.URI = uri
	return hrf
}

// WithStatus adiciona status ao HTTPRequestFields
func (hrf HTTPRequestFields) WithStatus(status int) HTTPRequestFields {
	hrf.Status = status
	return hrf
}

// WithLatency adiciona latency ao HTTPRequestFields
func (hrf HTTPRequestFields) WithLatency(latencyMs int64, duration string) HTTPRequestFields {
	hrf.LatencyMs = latencyMs
	hrf.Duration = duration
	return hrf
}

// WithUserAgent adiciona userAgent ao HTTPRequestFields
func (hrf HTTPRequestFields) WithUserAgent(userAgent string) HTTPRequestFields {
	hrf.UserAgent = userAgent
	return hrf
}

// WithIP adiciona ip ao HTTPRequestFields
func (hrf HTTPRequestFields) WithIP(ip string) HTTPRequestFields {
	hrf.IP = ip
	return hrf
}

// WithResponseSize adiciona responseSize ao HTTPRequestFields
func (hrf HTTPRequestFields) WithResponseSize(responseSize int64) HTTPRequestFields {
	hrf.ResponseSize = responseSize
	return hrf
}
