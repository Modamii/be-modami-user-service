package repository

import (
	"context"
	"time"

	"be-modami-user-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type outboxRepo struct {
	db *pgxpool.Pool
}

func NewOutboxRepository(db *pgxpool.Pool) *outboxRepo {
	return &outboxRepo{db: db}
}

func (r *outboxRepo) Create(ctx context.Context, topic, key string, payload []byte) error {
	id := uuid.New()
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		INSERT INTO outbox_events (id, topic, key, payload, status, created_at)
		VALUES ($1,$2,$3,$4,'pending',$5)`,
		id, topic, key, payload, time.Now(),
	)
	return err
}

func (r *outboxRepo) GetPending(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	rows, err := dbFromCtx(ctx, r.db).Query(ctx, `
		SELECT id, topic, key, payload, status, created_at, sent_at
		FROM outbox_events WHERE status='pending' ORDER BY created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.OutboxEvent
	for rows.Next() {
		e := &domain.OutboxEvent{}
		if err := rows.Scan(&e.ID, &e.Topic, &e.Key, &e.Payload, &e.Status, &e.CreatedAt, &e.SentAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *outboxRepo) MarkSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`UPDATE outbox_events SET status='sent', sent_at=$1 WHERE id=$2`, now, id)
	return err
}

func (r *outboxRepo) MarkFailed(ctx context.Context, id uuid.UUID) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`UPDATE outbox_events SET status='failed' WHERE id=$1`, id)
	return err
}

type processedEventRepo struct {
	db *pgxpool.Pool
}

func NewProcessedEventRepository(db *pgxpool.Pool) *processedEventRepo {
	return &processedEventRepo{db: db}
}

func (r *processedEventRepo) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id=$1)`, eventID,
	).Scan(&exists)
	return exists, err
}

func (r *processedEventRepo) MarkProcessed(ctx context.Context, eventID, topic string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO processed_events (event_id, topic, processed_at) VALUES ($1,$2,$3)
		ON CONFLICT (event_id) DO NOTHING`,
		eventID, topic, time.Now(),
	)
	return err
}

func (r *processedEventRepo) Cleanup(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec(ctx,
		`DELETE FROM processed_events WHERE processed_at < $1`, cutoff)
	return err
}
