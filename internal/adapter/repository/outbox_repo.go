package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type outboxRepo struct {
	db *pgxpool.Pool
}

func NewOutboxRepository(db *pgxpool.Pool) *outboxRepo {
	return &outboxRepo{db: db}
}

func (r *outboxRepo) Create(ctx context.Context, aggregateType, aggregateID, eventType string, payload []byte) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), aggregateType, aggregateID, eventType, payload, time.Now(),
	)
	return err
}

func (r *outboxRepo) Cleanup(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := dbFromCtx(ctx, r.db).Exec(ctx,
		`DELETE FROM outbox_events WHERE created_at < $1`, cutoff)
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
