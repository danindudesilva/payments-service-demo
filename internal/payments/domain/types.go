package domain

import "time"

type Money struct {
	Amount   int64
	Currency string
}

type NextActionType string

const (
	NextActionTypeNone     NextActionType = "none"
	NextActionTypeRedirect NextActionType = "redirect"
)

type NextAction struct {
	Type        NextActionType
	RedirectURL string
}

func NoNextAction() NextAction {
	return NextAction{
		Type: NextActionTypeNone,
	}
}

type ProviderDetails struct {
	ProviderName      string
	ProviderPaymentID string
	ClientSecret      string
}

type PaymentMethodDetails struct {
	Type string
}

type PaymentAttemptTimestamps struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}
