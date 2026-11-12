package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/order-service/internal/adapters/outbound/postgres"
	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
)

func TestOrderRepo_CreateAndGetByID(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	o := newOrder(nil)
	require.NoError(t, repo.Create(ctx, o))

	got, err := repo.GetByID(ctx, o.ID)
	require.NoError(t, err)
	require.Equal(t, o.ID, got.ID)
	require.Equal(t, o.UserID, got.UserID)
	require.Equal(t, o.MarketID, got.MarketID)
	require.Equal(t, o.Side, got.Side)
	require.Equal(t, o.Type, got.Type)
	require.Equal(t, o.Price, got.Price)
	require.Equal(t, o.Quantity, got.Quantity)
	require.Equal(t, o.Status, got.Status)
}

func TestOrderRepo_GetByID_NotFound(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, domain.OrderID("00000000-0000-0000-0000-000000000099"))
	require.ErrorIs(t, err, er.ErrOrderNotFound)
}

func TestOrderRepo_Cancel_Success(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	o := newOrder(nil)
	require.NoError(t, repo.Create(ctx, o))

	now := time.Now()
	cancelled, updated, err := repo.Cancel(ctx, o.ID, &o.UserID, now)
	require.NoError(t, err)
	require.True(t, updated)
	require.Equal(t, domain.OrderStatusCancelled, cancelled.Status)
	require.True(t, cancelled.UpdatedAt.After(o.UpdatedAt) || cancelled.UpdatedAt.Equal(now))

	// проверяем в БД
	got, err := repo.GetByID(ctx, o.ID)
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusCancelled, got.Status)
}

func TestOrderRepo_Cancel_NotFound(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	uid := domain.UserID("00000000-0000-0000-0000-000000000001")
	_, updated, err := repo.Cancel(ctx, "00000000-0000-0000-0000-000000000099", &uid, time.Now())
	require.NoError(t, err)
	require.False(t, updated)
}

func TestOrderRepo_Cancel_Forbidden(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	o := newOrder(nil)
	require.NoError(t, repo.Create(ctx, o))

	other := domain.UserID("00000000-0000-0000-0000-000000000002")
	_, updated, err := repo.Cancel(ctx, o.ID, &other, time.Now())
	require.NoError(t, err)
	require.False(t, updated) // чужой — не обновилось

	got, err := repo.GetByID(ctx, o.ID)
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusCreated, got.Status)
}

func TestOrderRepo_Cancel_AlreadyCancelled(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	o := newOrder(func(o *domain.Order) { o.Status = domain.OrderStatusCancelled })
	require.NoError(t, repo.Create(ctx, o))

	_, updated, err := repo.Cancel(ctx, o.ID, &o.UserID, time.Now())
	require.NoError(t, err)
	require.False(t, updated) // уже CANCELLED
}

func TestOrderRepo_Cancel_Admin(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	o := newOrder(nil)
	require.NoError(t, repo.Create(ctx, o))

	// userID = nil → без проверки владельца (админ)
	cancelled, updated, err := repo.Cancel(ctx, o.ID, nil, time.Now())
	require.NoError(t, err)
	require.True(t, updated)
	require.Equal(t, domain.OrderStatusCancelled, cancelled.Status)
}

func TestOrderRepo_ListByUser_Pagination(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	userID := domain.UserID("00000000-0000-0000-0000-000000000001")
	base := time.Now().Add(-time.Hour)

	// 3 заказа этого юзера + 1 чужой
	for i := 0; i < 3; i++ {
		o := newOrder(func(o *domain.Order) {
			o.UserID = userID
			o.CreatedAt = base.Add(time.Duration(i) * time.Minute)
			o.UpdatedAt = o.CreatedAt
		})
		require.NoError(t, repo.Create(ctx, o))
	}
	other := newOrder(nil)
	require.NoError(t, repo.Create(ctx, other))

	// первая страница: 2 заказа
	page1, err := repo.ListByUser(ctx, userID, 2, time.Time{}, "")
	require.NoError(t, err)
	require.Len(t, page1, 2)

	// вторая страница: 1 заказ (после последнего из page1)
	last := page1[len(page1)-1]
	page2, err := repo.ListByUser(ctx, userID, 2, last.CreatedAt, last.ID)
	require.NoError(t, err)
	require.Len(t, page2, 1)

	// все заказы — только этого юзера
	require.NotContains(t, orderIDs(page1, page2), other.ID)
}

func TestOrderRepo_ListAll_FilterAndPagination(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	user1 := domain.UserID("00000000-0000-0000-0000-000000000001")
	user2 := domain.UserID("00000000-0000-0000-0000-000000000002")
	base := time.Now().Add(-time.Hour)

	o1 := newOrder(func(o *domain.Order) { o.UserID = user1; o.CreatedAt = base; o.UpdatedAt = base })
	o2 := newOrder(func(o *domain.Order) {
		o.UserID = user2
		o.CreatedAt = base.Add(time.Minute)
		o.UpdatedAt = base.Add(time.Minute)
	})
	require.NoError(t, repo.Create(ctx, o1))
	require.NoError(t, repo.Create(ctx, o2))

	// все
	all, err := repo.ListAll(ctx, nil, 10, time.Time{}, "")
	require.NoError(t, err)
	require.Len(t, all, 2)

	// только user1
	onlyUser1, err := repo.ListAll(ctx, &user1, 10, time.Time{}, "")
	require.NoError(t, err)
	require.Len(t, onlyUser1, 1)
	require.Equal(t, o1.ID, onlyUser1[0].ID)

	// пагинация: page1 (1), потом page2
	page1, err := repo.ListAll(ctx, nil, 1, time.Time{}, "")
	require.NoError(t, err)
	require.Len(t, page1, 1)
	last := page1[0]
	page2, err := repo.ListAll(ctx, nil, 1, last.CreatedAt, last.ID)
	require.NoError(t, err)
	require.Len(t, page2, 1)
	require.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestOrderRepo_Create_RejectedStatus(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepo(pool)
	ctx := context.Background()

	o := newOrder(func(o *domain.Order) { o.Status = domain.OrderStatusRejected })
	require.NoError(t, repo.Create(ctx, o))

	got, err := repo.GetByID(ctx, o.ID)
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusRejected, got.Status)
}

func orderIDs(pages ...[]*domain.Order) []domain.OrderID {
	var ids []domain.OrderID
	for _, page := range pages {
		for _, o := range page {
			ids = append(ids, o.ID)
		}
	}
	return ids
}
