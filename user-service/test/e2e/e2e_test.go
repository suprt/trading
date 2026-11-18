package e2e_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	userv1 "github.com/suprt/trading/proto/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func authCtx(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func TestE2E_UserLifecycle(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// 1. Register
	_, err := client.Register(ctx, &userv1.RegisterRequest{
		Email:    "lifecycle@local",
		Username: "lifecycle_user",
		Password: "secret123",
	})
	require.NoError(t, err)

	// 2. Login
	loginResp, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "lifecycle@local",
		Password: "secret123",
	})
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.AccessToken)
	require.NotEmpty(t, loginResp.RefreshToken)
	require.Greater(t, loginResp.AccessExpiresIn, int64(0))

	// 3. GetSelfProfile — видим свой профиль
	auth := authCtx(ctx, loginResp.AccessToken)
	self, err := client.GetSelfProfile(auth, &userv1.GetSelfProfileRequest{})
	require.NoError(t, err)
	require.Equal(t, "lifecycle@local", self.User.Email)
	require.Equal(t, "lifecycle_user", self.User.Username)

	// 4. UpdateProfile — меняем username
	newUsername := "lifecycle_renamed"
	_, err = client.UpdateProfile(auth, &userv1.UpdateProfileRequest{
		Username: &newUsername,
	})
	require.NoError(t, err)

	// 5. GetSelfProfile — видим новый username
	self, err = client.GetSelfProfile(auth, &userv1.GetSelfProfileRequest{})
	require.NoError(t, err)
	require.Equal(t, "lifecycle_renamed", self.User.Username)

	// 6. ChangePassword
	_, err = client.ChangePassword(auth, &userv1.ChangePasswordRequest{
		CurrentPassword: "secret123",
		NewPassword:     "newsecret123",
	})
	require.NoError(t, err)

	// 7. Login со старым паролем → Unauthenticated
	_, err = client.Login(ctx, &userv1.LoginRequest{
		Email:    "lifecycle@local",
		Password: "secret123",
	})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	// 8. Login с новым паролем → успех
	loginResp2, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "lifecycle@local",
		Password: "newsecret123",
	})
	require.NoError(t, err)

	// 9. RefreshToken с новым refresh → успех (ротация)
	refreshResp, err := client.RefreshToken(ctx, &userv1.RefreshTokenRequest{
		RefreshToken: loginResp2.RefreshToken,
	})
	require.NoError(t, err)
	require.NotEmpty(t, refreshResp.AccessToken)
	require.NotEqual(t, loginResp2.RefreshToken, refreshResp.RefreshToken) // ротация

	// 10. RefreshToken со старым refresh → Unauthenticated (удалён после ротации)
	_, err = client.RefreshToken(ctx, &userv1.RefreshTokenRequest{
		RefreshToken: loginResp2.RefreshToken,
	})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestE2E_Security(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// 1. GetSelfProfile без токена → Unauthenticated
	_, err := client.GetSelfProfile(ctx, &userv1.GetSelfProfileRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	// 2. GetSelfProfile с битым токеном → Unauthenticated
	auth := authCtx(ctx, "invalid-token")
	_, err = client.GetSelfProfile(auth, &userv1.GetSelfProfileRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	// 3. Register с коротким паролем → InvalidArgument (protovalidate)
	_, err = client.Register(ctx, &userv1.RegisterRequest{
		Email:    "short@local",
		Username: "short_user",
		Password: "123",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// 4. Register с невалидным username (спецсимволы) → InvalidArgument
	_, err = client.Register(ctx, &userv1.RegisterRequest{
		Email:    "invalid@local",
		Username: "invalid user!",
		Password: "secret123",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// 5. GetUserByID обычным юзером → PermissionDenied
	_, err = client.Register(ctx, &userv1.RegisterRequest{
		Email: "normal@local", Username: "normal_user", Password: "secret123",
	})
	require.NoError(t, err)
	loginResp, err := client.Login(ctx, &userv1.LoginRequest{Email: "normal@local", Password: "secret123"})
	require.NoError(t, err)

	auth = authCtx(ctx, loginResp.AccessToken)
	_, err = client.GetUserByID(auth, &userv1.GetUserByIDRequest{UserId: loginResp.UserId})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestE2E_Conflicts(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Первый юзер
	_, err := client.Register(ctx, &userv1.RegisterRequest{
		Email: "conflict@local", Username: "conflict_user", Password: "secret123",
	})
	require.NoError(t, err)

	// Email занят
	_, err = client.Register(ctx, &userv1.RegisterRequest{
		Email: "conflict@local", Username: "different_user", Password: "secret123",
	})
	require.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.Contains(t, err.Error(), "email already taken")

	// Username занят
	_, err = client.Register(ctx, &userv1.RegisterRequest{
		Email: "different@local", Username: "conflict_user", Password: "secret123",
	})
	require.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.Contains(t, err.Error(), "username already taken")
}

func TestE2E_PublicProfile(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Создаём юзера
	regResp, err := client.Register(ctx, &userv1.RegisterRequest{
		Email: "public@local", Username: "public_user", Password: "secret123",
	})
	require.NoError(t, err)

	// GetPublicProfile без токена — публичный метод
	prof, err := client.GetPublicProfile(ctx, &userv1.GetPublicProfileRequest{
		UserId: regResp.UserId,
	})
	require.NoError(t, err)
	require.Equal(t, "public_user", prof.Profile.Username)
	require.NotEmpty(t, prof.Profile.CreatedAt)

	// Несуществующий user_id → NotFound
	_, err = client.GetPublicProfile(ctx, &userv1.GetPublicProfileRequest{
		UserId: "00000000-0000-0000-0000-000000000000",
	})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestE2E_AdminFlow(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// 1. Создаём обычного юзера (target)
	targetResp, err := client.Register(ctx, &userv1.RegisterRequest{
		Email:    "target@local",
		Username: "target_user",
		Password: "secret123",
	})
	require.NoError(t, err)
	targetID := targetResp.UserId

	// 2. Логинимся под админом (создан сидом)
	adminLogin, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "admin@test",
		Password: "admin12345",
	})
	require.NoError(t, err)
	adminAuth := authCtx(ctx, adminLogin.AccessToken)

	// 3. GetUserByID — админ читает профиль target
	userResp, err := client.GetUserByID(adminAuth, &userv1.GetUserByIDRequest{UserId: targetID})
	require.NoError(t, err)
	require.Equal(t, "target_user", userResp.User.Username)

	// 4. AdminUpdateProfile — меняем username target'а
	newUsername := "target_renamed"
	_, err = client.UpdateProfile(adminAuth, &userv1.UpdateProfileRequest{
		UserId:   &targetID,
		Username: &newUsername,
	})
	require.NoError(t, err)

	// 5. GetUserByID — видим новый username
	userResp, err = client.GetUserByID(adminAuth, &userv1.GetUserByIDRequest{UserId: targetID})
	require.NoError(t, err)
	require.Equal(t, "target_renamed", userResp.User.Username)

	// 6. AdminResetPassword — сбрасываем пароль target'а
	resetResp, err := client.AdminResetPassword(adminAuth, &userv1.AdminResetPasswordRequest{
		UserId: targetID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resetResp.TemporaryPassword)
	tempPassword := resetResp.TemporaryPassword

	// 7. Логин target'а со старым паролем → Unauthenticated
	_, err = client.Login(ctx, &userv1.LoginRequest{
		Email:    "target@local",
		Password: "secret123",
	})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	// 8. Логин target'а с новым временным паролем → успех
	targetLogin, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "target@local",
		Password: tempPassword,
	})
	require.NoError(t, err)

	// 9. AddRole — выдаём TRADER target'у
	_, err = client.AddRole(adminAuth, &userv1.AddRoleRequest{
		UserId: targetID,
		Role:   userv1.Role_ROLE_TRADER,
	})
	require.NoError(t, err)

	// 10. GetUserByID — видим TRADER в ролях
	userResp, err = client.GetUserByID(adminAuth, &userv1.GetUserByIDRequest{UserId: targetID})
	require.NoError(t, err)
	require.Contains(t, userResp.User.Roles, userv1.Role_ROLE_TRADER)

	// 11. RemoveRole — забираем TRADER
	_, err = client.RemoveRole(adminAuth, &userv1.RemoveRoleRequest{
		UserId: targetID,
		Role:   userv1.Role_ROLE_TRADER,
	})
	require.NoError(t, err)

	// 12. GetUserByID — TRADER больше нет
	userResp, err = client.GetUserByID(adminAuth, &userv1.GetUserByIDRequest{UserId: targetID})
	require.NoError(t, err)
	require.NotContains(t, userResp.User.Roles, userv1.Role_ROLE_TRADER)

	// 13. RefreshToken после RemoveRole → Unauthenticated
	// (RemoveRole инвалидирует сессии target'а)
	_, err = client.RefreshToken(ctx, &userv1.RefreshTokenRequest{
		RefreshToken: targetLogin.RefreshToken,
	})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func fakeIPContext(ctx context.Context, ip string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-forwarded-for", ip)
}

func TestE2E_RateLimit_Login(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := fakeIPContext(context.Background(), "192.0.2.1")

	// Лимит 5 в минуту на Login
	for i := 0; i < 5; i++ {
		_, err := client.Login(ctx, &userv1.LoginRequest{
			Email:    "nonexistent@local",
			Password: "wrongpass",
		})
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err), "request %d", i)
	}

	// 6-й — ResourceExhausted
	_, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "nonexistent@local",
		Password: "wrongpass",
	})
	require.Error(t, err)
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func TestE2E_RateLimit_Isolation(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := fakeIPContext(context.Background(), "192.0.2.2")

	// Исчерпываем лимит Login (5 запросов)
	for i := 0; i < 5; i++ {
		_, _ = client.Login(ctx, &userv1.LoginRequest{
			Email:    "x@local",
			Password: "wrongpass",
		})
	}

	// Login заблокирован
	_, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "x@local",
		Password: "wrongpass",
	})
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))

	// Register — работает (другой метод, другой счётчик)
	_, err = client.Register(ctx, &userv1.RegisterRequest{
		Email:    "new@local",
		Username: "new_user",
		Password: "secret123",
	})
	assert.Equal(t, codes.OK, status.Code(err))
}

func TestE2E_RateLimit_DifferentIPs(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	// IP 1 исчерпывает лимит
	ctx1 := fakeIPContext(context.Background(), "192.0.2.10")
	for i := 0; i < 5; i++ {
		_, _ = client.Login(ctx1, &userv1.LoginRequest{Email: "x@local", Password: "wrongpass"})
	}
	_, err := client.Login(ctx1, &userv1.LoginRequest{Email: "x@local", Password: "wrongpass"})
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))

	// IP 2 — свой счётчик, может логиниться
	ctx2 := fakeIPContext(context.Background(), "192.0.2.11")
	_, err = client.Login(ctx2, &userv1.LoginRequest{Email: "x@local", Password: "wrongpass"})
	assert.Equal(t, codes.Unauthenticated, status.Code(err)) // не ResourceExhausted
}

func TestE2E_RateLimit_RealIP(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "x-real-ip", "10.0.0.1")

	// Лимит 5 на Login
	for i := 0; i < 5; i++ {
		_, _ = client.Login(ctx, &userv1.LoginRequest{
			Email:    "x@local",
			Password: "wrongpassword",
		})
	}

	_, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "x@local",
		Password: "wrongpassword",
	})
	require.Error(t, err)
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func TestE2E_RateLimit_ForwardedFor(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	// Клиент прислал "1.1.1.1", прокси добавил "10.0.0.2"
	// Берём последний — 10.0.0.2
	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"x-forwarded-for", "1.1.1.1, 10.0.0.2")

	for i := 0; i < 5; i++ {
		_, _ = client.Login(ctx, &userv1.LoginRequest{
			Email:    "x@local",
			Password: "wrongpassword",
		})
	}

	_, err := client.Login(ctx, &userv1.LoginRequest{
		Email:    "x@local",
		Password: "wrongpassword",
	})
	require.Error(t, err)
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}
