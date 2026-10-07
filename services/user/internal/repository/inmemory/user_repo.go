package inmemory

import (
	"context"
	"sync"

	"github.com/suprt/trading/services/user/internal/domain"
)

type UserRepo struct {
	mu      sync.RWMutex
	byID    map[domain.UserID]*domain.User
	byEmail map[string]domain.UserID
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		byID:    make(map[domain.UserID]*domain.User),
		byEmail: make(map[string]domain.UserID),
	}
}

func (r *UserRepo) Create(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[u.ID]; ok {
		return domain.ErrDuplicateID
	}
	if _, ok := r.byEmail[u.Email]; ok {
		return domain.ErrEmailTaken
	}
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return nil
}

func (r *UserRepo) Update(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.byID[u.ID]
	if !ok {
		return domain.ErrUserNotFound
	}
	if old.Email != u.Email {
		delete(r.byEmail, old.Email)
		r.byEmail[u.Email] = u.ID
	}
	r.byID[u.ID] = u
	return nil
}

func (r *UserRepo) GetByID(_ context.Context, id domain.UserID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return r.byID[id], nil
}
