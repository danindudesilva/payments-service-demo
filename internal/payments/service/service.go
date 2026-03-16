package service

import (
	"context"
	"fmt"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
)

type Clock func() time.Time

type IDGenerator func() string

type Service struct {
	repo       domain.PaymentAttemptRepository
	gateway    domain.PaymentGateway
	now        Clock
	generateID IDGenerator
}

type CreatePaymentAttemptInput struct {
	OrderID     string
	Amount      int64
	Currency    string
	ReturnURL   string
	Description string
}

type CreatePaymentAttemptOutput struct {
	Attempt *domain.PaymentAttempt
}

func New(
	repo domain.PaymentAttemptRepository,
	gateway domain.PaymentGateway,
	now Clock,
	generateID IDGenerator,
) *Service {
	if now == nil {
		now = time.Now
	}

	return &Service{
		repo:       repo,
		gateway:    gateway,
		now:        now,
		generateID: generateID,
	}
}

func (s *Service) CreatePaymentAttempt(
	ctx context.Context,
	input CreatePaymentAttemptInput,
) (*CreatePaymentAttemptOutput, error) {
	if s.generateID == nil {
		return nil, fmt.Errorf("id generator must not be nil")
	}

	attemptID := s.generateID()
	now := s.now()

	attempt, err := domain.NewPaymentAttempt(
		attemptID,
		input.OrderID,
		input.ReturnURL,
		domain.Money{
			Amount:   input.Amount,
			Currency: input.Currency,
		},
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("create payment attempt: %w", err)
	}

	providerResult, err := s.gateway.CreatePayment(ctx, domain.CreateProviderPaymentRequest{
		AttemptID:   attempt.ID,
		OrderID:     attempt.OrderID,
		Money:       attempt.Money,
		ReturnURL:   input.ReturnURL,
		Description: input.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("create provider payment: %w", err)
	}

	err = attempt.LinkProvider(
		providerResult.ProviderName,
		providerResult.ProviderPaymentID,
		providerResult.ClientSecret,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("link provider payment: %w", err)
	}

	if err := applyProviderResult(attempt, providerResult, now); err != nil {
		return nil, fmt.Errorf("apply provider result: %w", err)
	}

	if err := s.repo.Save(ctx, attempt); err != nil {
		return nil, fmt.Errorf("save payment attempt: %w", err)
	}

	return &CreatePaymentAttemptOutput{
		Attempt: attempt,
	}, nil
}

func (s *Service) GetPaymentAttempt(ctx context.Context, attemptID string) (*domain.PaymentAttempt, error) {
	attempt, err := s.repo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, fmt.Errorf("get payment attempt: %w", err)
	}

	return attempt, nil
}

func (s *Service) ReconcilePaymentAttempt(ctx context.Context, attemptID string) (*domain.PaymentAttempt, error) {
	attempt, err := s.repo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, fmt.Errorf("get payment attempt: %w", err)
	}

	providerPaymentID, err := attempt.ProviderPaymentID()
	if err != nil {
		return nil, fmt.Errorf("get provider payment id: %w", err)
	}

	providerResult, err := s.gateway.GetPayment(ctx, providerPaymentID)
	if err != nil {
		return nil, fmt.Errorf("get provider payment: %w", err)
	}

	now := s.now()

	if err := applyProviderResult(attempt, providerResult, now); err != nil {
		return nil, fmt.Errorf("apply provider result: %w", err)
	}

	if err := s.repo.Save(ctx, attempt); err != nil {
		return nil, fmt.Errorf("save reconciled payment attempt: %w", err)
	}

	return attempt, nil
}

func applyProviderResult(
	attempt *domain.PaymentAttempt,
	result domain.CreateProviderPaymentResult,
	now time.Time,
) error {
	switch result.Status {
	case domain.PaymentStatusPending:
		return nil

	case domain.PaymentStatusRequiresAction:
		return attempt.MarkRequiresAction(result.NextAction, now)

	case domain.PaymentStatusProcessing:
		return attempt.MarkProcessing(now)

	case domain.PaymentStatusSucceeded:
		return attempt.MarkSucceeded(now)

	case domain.PaymentStatusFailed:
		return attempt.MarkFailed(domain.FailureReasonProviderReportedFailed, now)

	case domain.PaymentStatusCancelled:
		return attempt.MarkCancelled(now)

	default:
		return fmt.Errorf("unsupported provider status: %s", result.Status)
	}
}
