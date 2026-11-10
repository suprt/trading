package application

import (
	"context"
	"time"

	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
)

type CleanupSessionsUseCase struct {
	sessions outbound.SessionRepository
}

func NewCleanupSessionsUseCase(sessions outbound.SessionRepository) *CleanupSessionsUseCase {
	return &CleanupSessionsUseCase{sessions: sessions}
}

func (uc *CleanupSessionsUseCase) CleanupExpired(ctx context.Context) (int64, error) {
	return uc.sessions.DeleteExpired(ctx, time.Now())
}
