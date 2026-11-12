package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	shareddb "github.com/suprt/trading/shared/db"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

const orderColumns = `id, user_id, market_id, side, type, price, quantity, status, created_at, updated_at`

func scanOrder(row pgx.Row) (*domain.Order, error) {
	var o domain.Order
	var side, typ, status string
	err := row.Scan(
		&o.ID, &o.UserID, &o.MarketID, &side, &typ,
		&o.Price, &o.Quantity, &status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	o.Side = domain.OrderSide(side)
	o.Type = domain.OrderType(typ)
	o.Status = domain.OrderStatus(status)
	return &o, nil
}

func (r *OrderRepo) Create(ctx context.Context, o *domain.Order) error {
	db := shareddb.DBFromContext(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO orders (id, user_id, market_id, side, type, price, quantity, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		o.ID, o.UserID, o.MarketID, string(o.Side), string(o.Type),
		o.Price, o.Quantity, string(o.Status), o.CreatedAt, o.UpdatedAt)
	return err
}

func (r *OrderRepo) GetByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	db := shareddb.DBFromContext(ctx, r.pool)
	row := db.QueryRow(ctx, `SELECT `+orderColumns+` FROM orders WHERE id = $1`, id)
	o, err := scanOrder(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, er.ErrOrderNotFound
	}
	return o, err
}

func (r *OrderRepo) Cancel(ctx context.Context, id domain.OrderID, userID *domain.UserID, now time.Time) (*domain.Order, bool, error) {
	db := shareddb.DBFromContext(ctx, r.pool)

	row := db.QueryRow(ctx, `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
		  AND ($4::uuid IS NULL OR user_id = $4)
		  AND status = $5
		RETURNING `+orderColumns,
		string(domain.OrderStatusCancelled), now, id, userID, string(domain.OrderStatusCreated))

	o, err := scanOrder(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return o, true, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID domain.UserID, limit int,
	afterCreatedAt time.Time, afterID domain.OrderID) ([]*domain.Order, error) {
	db := shareddb.DBFromContext(ctx, r.pool)
	var rows pgx.Rows
	var err error

	if afterID == "" {
		rows, err = db.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			WHERE user_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2`, userID, limit)
	} else {
		rows, err = db.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			WHERE user_id = $1 AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4`, userID, afterCreatedAt, afterID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectOrders(rows)
}

func (r *OrderRepo) ListAll(ctx context.Context, userID *domain.UserID, limit int,
	afterCreatedAt time.Time, afterID domain.OrderID) ([]*domain.Order, error) {
	db := shareddb.DBFromContext(ctx, r.pool)
	var rows pgx.Rows
	var err error

	switch {
	case userID == nil && afterID == "":
		rows, err = db.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			ORDER BY created_at DESC, id DESC
			LIMIT $1`, limit)
	case userID == nil:
		rows, err = db.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			WHERE (created_at, id) < ($1, $2)
			ORDER BY created_at DESC, id DESC
			LIMIT $3`, afterCreatedAt, afterID, limit)
	case afterID == "":
		rows, err = db.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			WHERE user_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2`, *userID, limit)
	default: // userID != nil && afterID != ""
		rows, err = db.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			WHERE user_id = $1 AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4`, *userID, afterCreatedAt, afterID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectOrders(rows)
}

func collectOrders(rows pgx.Rows) ([]*domain.Order, error) {
	var orders []*domain.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
