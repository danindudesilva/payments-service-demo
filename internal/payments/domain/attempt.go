package domain

import (
	"fmt"
	"strings"
	"time"
)

type PaymentAttempt struct {
	ID            string
	OrderID       string
	Status        PaymentStatus
	Money         Money
	NextAction    NextAction
	Provider      ProviderDetails
	PaymentMethod PaymentMethodDetails
	Timestamps    PaymentAttemptTimestamps
	FailureReason string
}

func NewPaymentAttempt(
	id string,
	orderID string,
	money Money,
	now time.Time,
) (*PaymentAttempt, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("orderID must not be empty")
	}

	if err := validateMoney(money); err != nil {
		return nil, err
	}

	return &PaymentAttempt{
		ID:      id,
		OrderID: orderID,
		Status:  PaymentStatusPending,
		Money: Money{
			Amount:   money.Amount,
			Currency: strings.ToUpper(strings.TrimSpace(money.Currency)),
		},
		NextAction: NoNextAction(),
		Timestamps: PaymentAttemptTimestamps{
			CreatedAt: now.UTC(),
			UpdatedAt: now.UTC(),
		},
	}, nil
}

func (p *PaymentAttempt) LinkProvider(providerName, providerPaymentID, clientSecret string, now time.Time) error {
	if strings.TrimSpace(providerName) == "" {
		return fmt.Errorf("providerName must not be empty")
	}

	if strings.TrimSpace(providerPaymentID) == "" {
		return fmt.Errorf("providerPaymentID must not be empty")
	}

	p.Provider = ProviderDetails{
		ProviderName:      providerName,
		ProviderPaymentID: providerPaymentID,
		ClientSecret:      clientSecret,
	}
	p.touch(now)
	return nil
}

func (p *PaymentAttempt) MarkRequiresAction(action NextAction, now time.Time) error {
	if err := ensureValidNextAction(action); err != nil {
		return err
	}

	if err := p.transitionTo(PaymentStatusRequiresAction); err != nil {
		return err
	}

	p.NextAction = action
	p.FailureReason = ""
	p.touch(now)
	return nil
}

func (p *PaymentAttempt) MarkProcessing(now time.Time) error {
	if err := p.transitionTo(PaymentStatusProcessing); err != nil {
		return err
	}

	p.NextAction = NoNextAction()
	p.FailureReason = ""
	p.touch(now)
	return nil
}

func (p *PaymentAttempt) MarkSucceeded(now time.Time) error {
	if err := p.transitionTo(PaymentStatusSucceeded); err != nil {
		return err
	}

	p.NextAction = NoNextAction()
	p.FailureReason = ""
	p.complete(now)
	return nil
}

func (p *PaymentAttempt) MarkFailed(reason string, now time.Time) error {
	if strings.TrimSpace(reason) == "" {
		reason = "unknown"
	}

	if err := p.transitionTo(PaymentStatusFailed); err != nil {
		return err
	}

	p.NextAction = NoNextAction()
	p.FailureReason = reason
	p.complete(now)
	return nil
}

func (p *PaymentAttempt) MarkCancelled(now time.Time) error {
	if err := p.transitionTo(PaymentStatusCancelled); err != nil {
		return err
	}

	p.NextAction = NoNextAction()
	p.FailureReason = ""
	p.complete(now)
	return nil
}

func (p *PaymentAttempt) CanBeResumed() bool {
	return p.Status == PaymentStatusRequiresAction || p.Status == PaymentStatusProcessing
}

func (p *PaymentAttempt) ProviderPaymentID() (string, error) {
	if strings.TrimSpace(p.Provider.ProviderPaymentID) == "" {
		return "", ErrProviderNotLinked
	}

	return p.Provider.ProviderPaymentID, nil
}

func (p *PaymentAttempt) transitionTo(next PaymentStatus) error {
	if p.Status == next {
		return nil
	}

	switch p.Status {
	case PaymentStatusPending:
		switch next {
		case PaymentStatusRequiresAction, PaymentStatusProcessing, PaymentStatusSucceeded, PaymentStatusFailed, PaymentStatusCancelled:
			p.Status = next
			return nil
		}
	case PaymentStatusRequiresAction:
		switch next {
		case PaymentStatusProcessing, PaymentStatusSucceeded, PaymentStatusFailed, PaymentStatusCancelled:
			p.Status = next
			return nil
		}
	case PaymentStatusProcessing:
		switch next {
		case PaymentStatusSucceeded, PaymentStatusFailed, PaymentStatusCancelled:
			p.Status = next
			return nil
		}
	case PaymentStatusSucceeded, PaymentStatusFailed, PaymentStatusCancelled:
		return ErrInvalidTransition
	}

	return ErrInvalidTransition
}

func (p *PaymentAttempt) touch(now time.Time) {
	p.Timestamps.UpdatedAt = now.UTC()
}

func (p *PaymentAttempt) complete(now time.Time) {
	p.Timestamps.UpdatedAt = now.UTC()
	completedAt := now.UTC()
	p.Timestamps.CompletedAt = &completedAt
}

func validateMoney(m Money) error {
	if m.Amount <= 0 {
		return ErrInvalidMoney
	}

	if strings.TrimSpace(m.Currency) == "" {
		return ErrInvalidMoney
	}

	return nil
}

func ensureValidNextAction(action NextAction) error {
	switch action.Type {
	case NextActionTypeRedirect:
		if strings.TrimSpace(action.RedirectURL) == "" {
			return ErrInvalidNextAction
		}
		return nil

	case NextActionTypeNone:
		return ErrInvalidNextAction

	default:
		return ErrInvalidNextAction
	}
}
