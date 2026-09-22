package dto

// AccountLifecycleReasonDTO motivo obrigatório para block/delete da própria conta.
type AccountLifecycleReasonDTO struct {
	Reason string `json:"reason"`
}
