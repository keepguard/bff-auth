package dto

type ListDeviceBlacklistQuery struct {
	TenantID      string
	CorrelationID string
	Token         string
}

type AddDeviceBlacklistCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	DeviceID      string
	DeviceName    string
	Reason        string
}

type RemoveDeviceBlacklistCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	DeviceID      string
}

type SearchAdminDeviceBlacklistQuery struct {
	TenantID      string
	CorrelationID string
	Token         string
	QueryParams   map[string]string
}

type AdminAddDeviceBlacklistCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	UserID        string
	DeviceID      string
	DeviceName    string
	Reason        string
	ExpiresAt     string
}

type AdminRemoveDeviceBlacklistCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	DeviceID      string
	UserID        string
}

type ListTenantUserBlacklistQuery struct {
	TenantID      string
	CorrelationID string
	Token         string
	UserID        string
}
