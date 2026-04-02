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

type reviewRepo struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *reviewRepo {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) Create(ctx context.Context, review *domain.Review) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		INSERT INTO reviews (id, reviewer_id, reviewee_id, order_id, rating, comment, role, is_anonymous, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		review.ID, review.ReviewerID, review.RevieweeID, review.OrderID,
		review.Rating, review.Comment, review.Role, review.IsAnonymous,
		review.CreatedAt, review.UpdatedAt,
	)
	return err
}

func (r *reviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	row := dbFromCtx(ctx, r.db).QueryRow(ctx, `
		SELECT id, reviewer_id, reviewee_id, order_id, rating, comment, role, is_anonymous, created_at, updated_at
		FROM reviews WHERE id = $1`, id)
	return r.scanReview(row)
}

func (r *reviewRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Review, error) {
	row := dbFromCtx(ctx, r.db).QueryRow(ctx, `
		SELECT id, reviewer_id, reviewee_id, order_id, rating, comment, role, is_anonymous, created_at, updated_at
		FROM reviews WHERE order_id = $1`, orderID)
	return r.scanReview(row)
}

func (r *reviewRepo) scanReview(row pgx.Row) (*domain.Review, error) {
	rv := &domain.Review{}
	err := row.Scan(
		&rv.ID, &rv.ReviewerID, &rv.RevieweeID, &rv.OrderID,
		&rv.Rating, &rv.Comment, &rv.Role, &rv.IsAnonymous,
		&rv.CreatedAt, &rv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return rv, nil
}

func (r *reviewRepo) ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit int, cursor *time.Time) ([]*domain.Review, error) {
	var rows pgx.Rows
	var err error
	if cursor != nil {
		rows, err = dbFromCtx(ctx, r.db).Query(ctx, `
			SELECT id, reviewer_id, reviewee_id, order_id, rating, comment, role, is_anonymous, created_at, updated_at
			FROM reviews WHERE reviewee_id = $1 AND created_at < $2
			ORDER BY created_at DESC LIMIT $3`,
			revieweeID, cursor, limit,
		)
	} else {
		rows, err = dbFromCtx(ctx, r.db).Query(ctx, `
			SELECT id, reviewer_id, reviewee_id, order_id, rating, comment, role, is_anonymous, created_at, updated_at
			FROM reviews WHERE reviewee_id = $1
			ORDER BY created_at DESC LIMIT $2`,
			revieweeID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		rv := &domain.Review{}
		if err := rows.Scan(
			&rv.ID, &rv.ReviewerID, &rv.RevieweeID, &rv.OrderID,
			&rv.Rating, &rv.Comment, &rv.Role, &rv.IsAnonymous,
			&rv.CreatedAt, &rv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}

func (r *reviewRepo) GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.RatingSummary, error) {
	row := dbFromCtx(ctx, r.db).QueryRow(ctx, `
		SELECT user_id, avg_rating, total_reviews, count_1, count_2, count_3, count_4, count_5
		FROM rating_summaries WHERE user_id = $1`, userID)
	rs := &domain.RatingSummary{}
	err := row.Scan(&rs.UserID, &rs.AvgRating, &rs.TotalReviews, &rs.Count1, &rs.Count2, &rs.Count3, &rs.Count4, &rs.Count5)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.RatingSummary{UserID: userID}, nil
		}
		return nil, err
	}
	return rs, nil
}

func (r *reviewRepo) UpsertRatingSummary(ctx context.Context, userID uuid.UUID, rating int) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		INSERT INTO rating_summaries (user_id, avg_rating, total_reviews, count_1, count_2, count_3, count_4, count_5)
		VALUES ($1, $2, 1, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			total_reviews = rating_summaries.total_reviews + 1,
			count_1 = rating_summaries.count_1 + CASE WHEN $2::int = 1 THEN 1 ELSE 0 END,
			count_2 = rating_summaries.count_2 + CASE WHEN $2::int = 2 THEN 1 ELSE 0 END,
			count_3 = rating_summaries.count_3 + CASE WHEN $2::int = 3 THEN 1 ELSE 0 END,
			count_4 = rating_summaries.count_4 + CASE WHEN $2::int = 4 THEN 1 ELSE 0 END,
			count_5 = rating_summaries.count_5 + CASE WHEN $2::int = 5 THEN 1 ELSE 0 END,
			avg_rating = (
				(rating_summaries.avg_rating * rating_summaries.total_reviews + $2::numeric) /
				(rating_summaries.total_reviews + 1)
			)`,
		userID, rating,
		boolToInt(rating == 1), boolToInt(rating == 2), boolToInt(rating == 3),
		boolToInt(rating == 4), boolToInt(rating == 5),
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
