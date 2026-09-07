package dto

type ListUserSessionsQuery struct {
	TenantID      string
	CorrelationID string
	Token         string
	DeviceID      string
}

type RevokeSessionCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	DeviceID      string
}

type RevokeAllOtherSessionsCommand struct {
	TenantID        string
	CorrelationID   string
	Token           string
	CurrentDeviceID string
}

type ListTenantUserSessionsQuery struct {
	TenantID      string
	CorrelationID string
	Token         string
	UserID        string
}

type RevokeTenantUserSessionCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	UserID        string
	DeviceID      string
}

type SearchTenantSessionsQuery struct {
	TenantID      string
	CorrelationID string
	Token         string
	QueryParams   map[string]string
}
