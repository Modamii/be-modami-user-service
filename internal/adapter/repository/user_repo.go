package repository

import (
	"context"
	"errors"
	"time"

	apperror "be-modami-user-service/internal/apperror"
	"be-modami-user-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, keycloak_id, email, username, full_name, phone, avatar_url, cover_url, bio, gender,
			date_of_birth, role, status, email_verified, trust_score, follower_count, following_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, query,
		user.ID, user.KeycloakID, user.Email, user.UserName, user.FullName, user.Phone,
		user.AvatarURL, user.CoverURL, user.Bio, user.Gender,
		user.DateOfBirth, user.Role, user.Status, user.EmailVerified,
		user.TrustScore, user.FollowerCount, user.FollowingCount,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, full_name, phone, avatar_url, cover_url, bio, gender,
			date_of_birth, role, status, email_verified, trust_score, follower_count, following_count,
			created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`
	return r.scanUser(dbFromCtx(ctx, r.db).QueryRow(ctx, query, id))
}

func (r *userRepo) GetByKeycloakID(ctx context.Context, keycloakID string) (*domain.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, full_name, phone, avatar_url, cover_url, bio, gender,
			date_of_birth, role, status, email_verified, trust_score, follower_count, following_count,
			created_at, updated_at, deleted_at
		FROM users WHERE keycloak_id = $1 AND deleted_at IS NULL`
	return r.scanUser(dbFromCtx(ctx, r.db).QueryRow(ctx, query, keycloakID))
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, full_name, phone, avatar_url, cover_url, bio, gender,
			date_of_birth, role, status, email_verified, trust_score, follower_count, following_count,
			created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`
	return r.scanUser(dbFromCtx(ctx, r.db).QueryRow(ctx, query, email))
}

func (r *userRepo) scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.KeycloakID, &u.Email, &u.UserName, &u.FullName, &u.Phone,
		&u.AvatarURL, &u.CoverURL, &u.Bio, &u.Gender,
		&u.DateOfBirth, &u.Role, &u.Status, &u.EmailVerified,
		&u.TrustScore, &u.FollowerCount, &u.FollowingCount,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET full_name=$1, phone=$2, avatar_url=$3, cover_url=$4, bio=$5, gender=$6,
			date_of_birth=$7, updated_at=$8
		WHERE id=$9 AND deleted_at IS NULL`
	ct, err := dbFromCtx(ctx, r.db).Exec(ctx, query,
		user.FullName, user.Phone, user.AvatarURL, user.CoverURL,
		user.Bio, user.Gender, user.DateOfBirth, time.Now(), user.ID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *userRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at=$1, status='inactive', updated_at=$1 WHERE id=$2 AND deleted_at IS NULL`
	ct, err := dbFromCtx(ctx, r.db).Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *userRepo) Search(ctx context.Context, query string, limit int, cursor *time.Time) ([]*domain.User, error) {
	var rows pgx.Rows
	var err error
	if cursor != nil {
		sql := `
			SELECT id, keycloak_id, email, full_name, phone, avatar_url, cover_url, bio, gender,
				date_of_birth, role, status, email_verified, trust_score, follower_count, following_count,
				created_at, updated_at, deleted_at
			FROM users
			WHERE deleted_at IS NULL
				AND (full_name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
				AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3`
		rows, err = dbFromCtx(ctx, r.db).Query(ctx, sql, query, cursor, limit)
	} else {
		sql := `
			SELECT id, keycloak_id, email, full_name, phone, avatar_url, cover_url, bio, gender,
				date_of_birth, role, status, email_verified, trust_score, follower_count, following_count,
				created_at, updated_at, deleted_at
			FROM users
			WHERE deleted_at IS NULL
				AND (full_name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
			ORDER BY created_at DESC
			LIMIT $2`
		rows, err = dbFromCtx(ctx, r.db).Query(ctx, sql, query, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanUsers(rows)
}

func (r *userRepo) scanUsers(rows pgx.Rows) ([]*domain.User, error) {
	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		err := rows.Scan(
			&u.ID, &u.KeycloakID, &u.Email, &u.UserName, &u.FullName, &u.Phone,
			&u.AvatarURL, &u.CoverURL, &u.Bio, &u.Gender,
			&u.DateOfBirth, &u.Role, &u.Status, &u.EmailVerified,
			&u.TrustScore, &u.FollowerCount, &u.FollowingCount,
			&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepo) UpdateTrustScore(ctx context.Context, userID uuid.UUID, score float64) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`UPDATE users SET trust_score=$1, updated_at=$2 WHERE id=$3`,
		score, time.Now(), userID,
	)
	return err
}

func (r *userRepo) UpdateRole(ctx context.Context, userID uuid.UUID, role domain.UserRole) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`UPDATE users SET role=$1, updated_at=$2 WHERE id=$3`,
		role, time.Now(), userID,
	)
	return err
}

func (r *userRepo) UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`UPDATE users SET status=$1, updated_at=$2 WHERE id=$3`,
		status, time.Now(), userID,
	)
	return err
}

func (r *userRepo) IncrFollowerCount(ctx context.Context, userID uuid.UUID, delta int) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`UPDATE users SET follower_count = follower_count + $1, updated_at=$2 WHERE id=$3`,
		delta, time.Now(), userID,
	)
	return err
}

func (r *userRepo) IncrFollowingCount(ctx context.Context, userID uuid.UUID, delta int) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`UPDATE users SET following_count = following_count + $1, updated_at=$2 WHERE id=$3`,
		delta, time.Now(), userID,
	)
	return err
}
