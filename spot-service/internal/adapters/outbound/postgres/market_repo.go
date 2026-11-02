package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	shareddb "github.com/suprt/trading/shared/db"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

type MarketRepo struct {
	pool *pgxpool.Pool
}

func NewMarketRepo(pool *pgxpool.Pool) *MarketRepo {
	return &MarketRepo{pool: pool}
}

func scanMarket(row pgx.Row) (*domain.Market, error) {
	var m domain.Market
	var roles []string
	var deletedAt *time.Time

	err := row.Scan(
		&m.ID, &m.Name, &roles, &m.Enabled, &deletedAt, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}

	m.DeletedAt = deletedAt
	m.RequiredRoles = make([]domain.Role, len(roles))
	for i, role := range roles {
		m.RequiredRoles[i] = domain.Role(role)
	}
	return &m, nil
}

func (r *MarketRepo) List(ctx context.Context) ([]*domain.Market, error) {
	db := shareddb.DBFromContext(ctx, r.pool)

	rows, err := db.Query(ctx, `
			SELECT id, name, required_roles, enabled, deleted_at, created_at, updated_at
			FROM markets
			WHERE enabled = TRUE AND deleted_at IS NULL
			ORDER BY created_at DESC
			`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []*domain.Market
	for rows.Next() {
		m, err := scanMarket(rows)
		if err != nil {
			return nil, err
		}
		markets = append(markets, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *MarketRepo) GetByID(ctx context.Context, id domain.MarketID) (*domain.Market, error) {
	db := shareddb.DBFromContext(ctx, r.pool)
	row := db.QueryRow(ctx, `
			SELECT id, name, required_roles, enabled, deleted_at, created_at, updated_at
			FROM markets
			WHERE id = $1
			`, id)
	m, err := scanMarket(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, er.ErrMarketNotFound
	}
	return m, err
}

func (r *MarketRepo) GetByName(ctx context.Context, name string) (*domain.Market, error) {
	db := shareddb.DBFromContext(ctx, r.pool)
	row := db.QueryRow(ctx, `
			SELECT id, name, required_roles, enabled, deleted_at, created_at, updated_at
			FROM markets
			WHERE name = $1
			`, name)

	m, err := scanMarket(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, er.ErrMarketNotFound
	}
	return m, err
}

func (r *MarketRepo) Create(ctx context.Context, m *domain.Market) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	roles := make([]string, len(m.RequiredRoles))
	for i, role := range m.RequiredRoles {
		roles[i] = string(role)
	}

	_, err := db.Exec(ctx, `
			INSERT INTO markets (id, name, required_roles, enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.Name, m.RequiredRoles, m.Enabled, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *MarketRepo) UpdateDetails(ctx context.Context, m *domain.Market) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	res, err := db.Exec(ctx, `
			UPDATE markets
			SET name = $1, enabled = $2, updated_at = $3
			WHERE id = $4`,
		m.Name, m.Enabled, m.UpdatedAt, m.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return er.ErrMarketNotFound
	}
	return nil
}

func (r *MarketRepo) UpdateRoles(ctx context.Context, m *domain.Market) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	roles := make([]string, len(m.RequiredRoles))
	for i, role := range m.RequiredRoles {
		roles[i] = string(role)
	}

	res, err := db.Exec(ctx, `
			UPDATE markets
			SET required_roles = $1, updated_at = $2
			WHERE id = $3`,
		m.RequiredRoles, m.UpdatedAt, m.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return er.ErrMarketNotFound
	}
	return nil
}

func (r *MarketRepo) SoftDelete(ctx context.Context, id domain.MarketID, deletedAt time.Time) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	res, err := db.Exec(ctx, `
			UPDATE markets
			SET deleted_at = $1, updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL`,
		deletedAt, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return er.ErrMarketNotFound
	}
	return nil
}
