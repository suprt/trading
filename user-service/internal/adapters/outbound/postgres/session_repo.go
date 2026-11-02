package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	shareddb "github.com/suprt/trading/shared/db"
	"github.com/suprt/trading/user-service/internal/core/domain"
	er "github.com/suprt/trading/user-service/internal/core/errors"
)

type SessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

func (r *SessionRepo) Save(ctx context.Context, s *domain.RefreshSession) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	_, err := db.Exec(ctx, `
				INSERT INTO sessions (token, user_id, expires_at, issued_at)
				VALUES ($1, $2, $3, $4)`,
		s.Token, s.UserID, s.ExpiresAt, s.IssuedAt)
	return err
}

func (r *SessionRepo) GetByToken(ctx context.Context, t domain.RefreshToken) (*domain.RefreshSession, error) {
	var s domain.RefreshSession
	db := shareddb.DBFromContext(ctx, r.pool)
	err := db.QueryRow(ctx, `
				SELECT token, user_id, expires_at, issued_at
				FROM sessions WHERE token = $1`,
		t).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.IssuedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, er.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepo) Delete(ctx context.Context, t domain.RefreshToken) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, t)
	return err
}

func (r *SessionRepo) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}
