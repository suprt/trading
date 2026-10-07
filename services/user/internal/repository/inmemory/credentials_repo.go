package inmemory

import (
	"context"
	"sync"

	"github.com/suprt/ITK-proj/services/user/internal/domain"
)

type CredentialsRepo struct {
	mu   sync.RWMutex
	data map[domain.UserID]*domain.Credentials
}

func NewCredentialsRepo() *CredentialsRepo {
	return &CredentialsRepo{data: make(map[domain.UserID]*domain.Credentials)}
}

func (r *CredentialsRepo) Save(_ context.Context, c *domain.Credentials) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[c.UserID] = c
	return nil
}

func (r *CredentialsRepo) GetByID(_ context.Context, userID domain.UserID) (*domain.Credentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.data[userID]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return c, nil

}
