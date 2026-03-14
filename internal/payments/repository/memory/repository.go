package memory

import (
	"context"
	"sync"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
)

type Repository struct {
	mu                  sync.RWMutex
	attemptsByID        map[string]*domain.PaymentAttempt
	attemptIDByProvider map[string]string
}

func NewRepository() *Repository {
	return &Repository{
		attemptsByID:        make(map[string]*domain.PaymentAttempt),
		attemptIDByProvider: make(map[string]string),
	}
}

func (r *Repository) Save(_ context.Context, attempt *domain.PaymentAttempt) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cloned := cloneAttempt(attempt)
	r.attemptsByID[attempt.ID] = cloned

	if attempt.Provider.ProviderPaymentID != "" {
		r.attemptIDByProvider[attempt.Provider.ProviderPaymentID] = attempt.ID
	}

	return nil
}

func (r *Repository) GetByID(_ context.Context, id string) (*domain.PaymentAttempt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	attempt, ok := r.attemptsByID[id]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}

	return cloneAttempt(attempt), nil
}

func (r *Repository) GetByProviderPaymentID(_ context.Context, providerPaymentID string) (*domain.PaymentAttempt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	attemptID, ok := r.attemptIDByProvider[providerPaymentID]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}

	attempt, ok := r.attemptsByID[attemptID]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}

	return cloneAttempt(attempt), nil
}

func cloneAttempt(attempt *domain.PaymentAttempt) *domain.PaymentAttempt {
	if attempt == nil {
		return nil
	}

	cloned := *attempt

	if attempt.Timestamps.CompletedAt != nil {
		completedAt := *attempt.Timestamps.CompletedAt
		cloned.Timestamps.CompletedAt = &completedAt
	}

	return &cloned
}
