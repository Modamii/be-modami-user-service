package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/pkg/apperror"
)

type followRepo struct {
	db *pgxpool.Pool
}

func NewFollowRepository(db *pgxpool.Pool) *followRepo {
	return &followRepo{db: db}
}

func (r *followRepo) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO user_follows (follower_id, following_id, created_at) VALUES ($1, $2, $3)`,
		followerID, followingID, time.Now(),
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE users SET follower_count = follower_count + 1, updated_at = $1 WHERE id = $2`,
		time.Now(), followingID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE users SET following_count = following_count + 1, updated_at = $1 WHERE id = $2`,
		time.Now(), followerID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *followRepo) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx,
		`DELETE FROM user_follows WHERE follower_id = $1 AND following_id = $2`,
		followerID, followingID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperror.ErrNotFollowing
	}

	_, err = tx.Exec(ctx,
		`UPDATE users SET follower_count = GREATEST(follower_count - 1, 0), updated_at = $1 WHERE id = $2`,
		time.Now(), followingID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE users SET following_count = GREATEST(following_count - 1, 0), updated_at = $1 WHERE id = $2`,
		time.Now(), followerID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *followRepo) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID,
	).Scan(&exists)
	return exists, err
}

func (r *followRepo) GetFollowers(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]*domain.FollowUser, error) {
	var rows pgx.Rows
	var err error
	if cursor != nil {
		rows, err = r.db.Query(ctx, `
			SELECT u.id, u.full_name, u.avatar_url, f.created_at
			FROM user_follows f
			JOIN users u ON u.id = f.follower_id
			WHERE f.following_id = $1 AND f.created_at < $2 AND u.deleted_at IS NULL
			ORDER BY f.created_at DESC
			LIMIT $3`,
			userID, cursor, limit,
		)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT u.id, u.full_name, u.avatar_url, f.created_at
			FROM user_follows f
			JOIN users u ON u.id = f.follower_id
			WHERE f.following_id = $1 AND u.deleted_at IS NULL
			ORDER BY f.created_at DESC
			LIMIT $2`,
			userID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanFollowUsers(rows)
}

func (r *followRepo) GetFollowing(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]*domain.FollowUser, error) {
	var rows pgx.Rows
	var err error
	if cursor != nil {
		rows, err = r.db.Query(ctx, `
			SELECT u.id, u.full_name, u.avatar_url, f.created_at
			FROM user_follows f
			JOIN users u ON u.id = f.following_id
			WHERE f.follower_id = $1 AND f.created_at < $2 AND u.deleted_at IS NULL
			ORDER BY f.created_at DESC
			LIMIT $3`,
			userID, cursor, limit,
		)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT u.id, u.full_name, u.avatar_url, f.created_at
			FROM user_follows f
			JOIN users u ON u.id = f.following_id
			WHERE f.follower_id = $1 AND u.deleted_at IS NULL
			ORDER BY f.created_at DESC
			LIMIT $2`,
			userID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanFollowUsers(rows)
}

func (r *followRepo) scanFollowUsers(rows pgx.Rows) ([]*domain.FollowUser, error) {
	var users []*domain.FollowUser
	for rows.Next() {
		u := &domain.FollowUser{}
		if err := rows.Scan(&u.ID, &u.FullName, &u.AvatarURL, &u.FollowedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

