package inmemory

import (
	"context"
	"sync"

	"github.com/suprt/ITK-proj/services/user/internal/domain"
)

type SessionRepo struct {
	mu   sync.RWMutex
	data map[domain.RefreshToken]*domain.RefreshSession
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{data: make(map[domain.RefreshToken]*domain.RefreshSession)}
}

func (r *SessionRepo) Save(_ context.Context, s *domain.RefreshSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.Token] = s
	return nil
}

func (r *SessionRepo) GetByToken(_ context.Context, t domain.RefreshToken) (*domain.RefreshSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.data[t]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	return session, nil
}

func (r *SessionRepo) Delete(_ context.Context, t domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, t)
	return nil
}
