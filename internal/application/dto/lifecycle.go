package dto

type AccountLifecycleCommand struct {
	TenantID      string
	Token         string
	CodeUser      string
	Reason        string
	CorrelationID string
}
