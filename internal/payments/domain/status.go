package domain

type PaymentStatus string

const (
	PaymentStatusPending        PaymentStatus = "pending"
	PaymentStatusRequiresAction PaymentStatus = "requires_action"
	PaymentStatusProcessing     PaymentStatus = "processing"
	PaymentStatusSucceeded      PaymentStatus = "succeeded"
	PaymentStatusFailed         PaymentStatus = "failed"
	PaymentStatusCancelled      PaymentStatus = "cancelled"
)

func (s PaymentStatus) IsTerminal() bool {
	switch s {
	case PaymentStatusSucceeded, PaymentStatusFailed, PaymentStatusCancelled:
		return true

	default:
		return false
	}
}
