package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/suprt/trading/user-service/internal/application"
	"github.com/suprt/trading/user-service/internal/core/auth"
	"github.com/suprt/trading/user-service/internal/core/domain"
	"github.com/suprt/trading/user-service/internal/core/dto"
	er "github.com/suprt/trading/user-service/internal/core/errors"
)

func setupGetUserByIDTest() (*mockUserRepo, *application.GetUserByIDUseCase) {
	users := newMockUserRepo()
	uc := application.NewGetUserByIDUseCase(users)
	return users, uc
}

func TestGetUserByID_Success(t *testing.T) {
	users, uc := setupGetUserByIDTest()

	adminID := domain.UserID("admin")
	targetID := domain.UserID("user")
	createUser(t, users, adminID, []domain.Role{domain.RoleAdmin})
	createUser(t, users, targetID, []domain.Role{domain.RoleUser})

	ctx := auth.WithUserID(context.Background(), adminID)

	out, err := uc.GetUserByID(ctx, dto.GetUserByIDInput{UserID: targetID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.ID != targetID {
		t.Fatalf("expected user %s, got %s", targetID, out.User.ID)
	}
}

func TestGetUserByID_Forbidden(t *testing.T) {
	users, uc := setupGetUserByIDTest()

	userID := domain.UserID("user")
	targetID := domain.UserID("target")
	createUser(t, users, userID, []domain.Role{domain.RoleUser})
	createUser(t, users, targetID, []domain.Role{domain.RoleUser})

	ctx := auth.WithUserID(context.Background(), userID)

	_, err := uc.GetUserByID(ctx, dto.GetUserByIDInput{UserID: targetID})
	if !errors.Is(err, er.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
