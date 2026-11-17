package application

import (
	"context"
	"time"

	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
)

type CleanupSessionsUseCase struct {
	sessions outbound.SessionRepository
	uow      outbound.UnitOfWork
}

func NewCleanupSessionsUseCase(sessions outbound.SessionRepository, uow outbound.UnitOfWork) *CleanupSessionsUseCase {
	return &CleanupSessionsUseCase{sessions: sessions, uow: uow}
}

func (uc *CleanupSessionsUseCase) CleanupExpired(ctx context.Context) (int64, error) {
	var n int64
	err := uc.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		n, err = uc.sessions.DeleteExpired(ctx, time.Now())
		return err
	})
	return n, err
}
