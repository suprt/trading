package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	shareddb "github.com/suprt/trading/shared/db"
	"github.com/suprt/trading/user-service/internal/core/domain"
	er "github.com/suprt/trading/user-service/internal/core/errors"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) (err error) {
	db := shareddb.DBFromContext(ctx, r.pool)

	_, err = db.Exec(ctx,
		`INSERT INTO users(id, email, username, roles_version, created_at, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.Email, u.Username, u.RolesVersion, u.CreatedAt, u.UpdatedAt)

	if err != nil {
		if uv, ok := shareddb.AsUniqueViolation(err); ok {
			switch uv.Constraint {
			case "users_email_unique":
				return er.ErrEmailTaken
			case "users_username_unique":
				return er.ErrUsernameTaken
			}
		}
		return err
	}

	for _, role := range u.Roles {
		_, err = db.Exec(ctx, `
		INSERT INTO user_roles(user_id, role) VALUES ($1, $2)`,
			u.ID, string(role))

		if err != nil {
			return err
		}
	}
	return nil
}
func (r *UserRepo) UpdateProfile(ctx context.Context, u *domain.User) (err error) {
	db := shareddb.DBFromContext(ctx, r.pool)

	res, err := db.Exec(ctx, `
		UPDATE users 
		SET email = $1, username = $2, updated_at = $3
		WHERE id = $4`,
		u.Email, u.Username, u.UpdatedAt, u.ID)
	if err != nil {
		if uv, ok := shareddb.AsUniqueViolation(err); ok {
			switch uv.Constraint {
			case "users_email_unique":
				return er.ErrEmailTaken
			case "users_username_unique":
				return er.ErrUsernameTaken
			}
		}
		return err
	}
	n := res.RowsAffected()
	if n == 0 {
		return er.ErrUserNotFound
	}
	return nil
}
func (r *UserRepo) UpdateRoles(ctx context.Context, u *domain.User) (err error) {
	db := shareddb.DBFromContext(ctx, r.pool)

	res, err := db.Exec(ctx, `
		UPDATE users 
		SET roles_version = $1, updated_at = $2
		WHERE id = $3`,
		u.RolesVersion, u.UpdatedAt, u.ID)
	if err != nil {
		return err
	}
	n := res.RowsAffected()
	if n == 0 {
		return er.ErrUserNotFound
	}
	//TODO: разделить update
	_, err = db.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, u.ID)
	if err != nil {
		return err
	}

	for _, role := range u.Roles {
		_, err = db.Exec(ctx, `
		INSERT INTO user_roles(user_id, role) VALUES ($1, $2)`,
			u.ID, string(role))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	var u domain.User
	var roles []string
	db := shareddb.DBFromContext(ctx, r.pool)
	err := db.QueryRow(ctx, `
 			SELECT u.id, u.email, u.username, u.roles_version, u.created_at, u.updated_at,
 			COALESCE(array_agg(r.role) FILTER ( WHERE r.role IS NOT NULL), '{}') AS roles
 			FROM users u 
 			LEFT JOIN user_roles r ON r.user_id = u.id
 			WHERE u.id = $1
 			GROUP BY u.id`, id).Scan(
		&u.ID, &u.Email, &u.Username, &u.RolesVersion, &u.CreatedAt, &u.UpdatedAt, &roles)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, er.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u.Roles = make([]domain.Role, len(roles))
	for i, role := range roles {
		u.Roles[i] = domain.Role(role)
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	var roles []string
	db := shareddb.DBFromContext(ctx, r.pool)
	err := db.QueryRow(ctx, `
 			SELECT u.id, u.email, u.username, u.roles_version, u.created_at, u.updated_at,
 			COALESCE(array_agg(r.role) FILTER ( WHERE r.role IS NOT NULL), '{}') AS roles
 			FROM users u 
 			LEFT JOIN user_roles r ON r.user_id = u.id
 			WHERE u.email = $1
 			GROUP BY u.id`, email).Scan(
		&u.ID, &u.Email, &u.Username, &u.RolesVersion, &u.CreatedAt, &u.UpdatedAt, &roles)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, er.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u.Roles = make([]domain.Role, len(roles))
	for i, role := range roles {
		u.Roles[i] = domain.Role(role)
	}
	return &u, nil
}

func (r *UserRepo) List(ctx context.Context, limit int, afterCreatedAt time.Time, afterID domain.UserID) ([]*domain.User, error) {
	db := shareddb.DBFromContext(ctx, r.pool)
	var rows pgx.Rows
	var err error

	if afterID == "" {
		rows, err = db.Query(ctx, `
 			SELECT u.id, u.email, u.username, u.roles_version, u.created_at, u.updated_at,
			COALESCE(array_agg(r.role) FILTER (WHERE r.role IS NOT NULL), '{}') AS roles
			FROM users u
			LEFT JOIN user_roles r ON r.user_id = u.id
			GROUP BY u.id, u.created_at
			ORDER BY u.created_at, u.id
			LIMIT $1`, limit)
	} else {
		rows, err = db.Query(ctx, `
 			SELECT u.id, u.email, u.username, u.roles_version, u.created_at, u.updated_at,
			COALESCE(array_agg(r.role) FILTER (WHERE r.role IS NOT NULL), '{}') AS roles
			FROM users u
			LEFT JOIN user_roles r ON r.user_id = u.id
			WHERE (u.created_at, u.id) > ($1, $2)
			GROUP BY u.id, u.created_at
			ORDER BY u.created_at, u.id
			LIMIT $3`, afterCreatedAt, afterID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var roles []string
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.RolesVersion, &u.CreatedAt, &u.UpdatedAt, &roles); err != nil {
			return nil, err
		}
		u.Roles = make([]domain.Role, len(roles))
		for i, role := range roles {
			u.Roles[i] = domain.Role(role)
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}
