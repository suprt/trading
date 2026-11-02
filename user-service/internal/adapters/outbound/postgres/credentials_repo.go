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

type CredentialRepo struct {
	pool *pgxpool.Pool
}

func NewCredentialsRepo(pool *pgxpool.Pool) *CredentialRepo {
	return &CredentialRepo{pool: pool}
}

func (r *CredentialRepo) Save(ctx context.Context, c *domain.Credentials) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	_, err := db.Exec(ctx, `
       	INSERT INTO credentials (user_id, password_hash)
       	VALUES ($1, $2)
       	ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash
       	`, c.UserID, c.PasswordHash)
	return err
}

func (r *CredentialRepo) GetByID(ctx context.Context, userID domain.UserID) (*domain.Credentials, error) {
	var c domain.Credentials
	db := shareddb.DBFromContext(ctx, r.pool)
	err := db.QueryRow(ctx, `
			SELECT user_id, password_hash FROM credentials WHERE user_id = $1`,
		userID).Scan(&c.UserID, &c.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, er.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil

}
