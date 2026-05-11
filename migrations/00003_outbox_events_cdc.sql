-- +goose Up
-- Migrate outbox_events to Debezium CDC Outbox EventRouter schema.
-- Drops polling columns (status, sent_at, topic, key) and adds
-- aggregate_type, aggregate_id, event_type for Debezium routing.

ALTER TABLE outbox_events
    ADD COLUMN aggregate_type VARCHAR(255),
    ADD COLUMN aggregate_id   VARCHAR(255),
    ADD COLUMN event_type     VARCHAR(255);

UPDATE outbox_events
SET aggregate_type = 'user',
    aggregate_id   = key,
    event_type     = 'Unknown';

ALTER TABLE outbox_events
    ALTER COLUMN aggregate_type SET NOT NULL,
    ALTER COLUMN aggregate_id   SET NOT NULL,
    ALTER COLUMN event_type     SET NOT NULL,
    DROP COLUMN topic,
    DROP COLUMN key,
    DROP COLUMN status,
    DROP COLUMN sent_at;

DROP INDEX IF EXISTS idx_outbox_events_status;
CREATE INDEX idx_outbox_events_created_at ON outbox_events(created_at);

-- Debezium pgoutput plugin requires a publication
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'user_service_outbox_pub') THEN
        CREATE PUBLICATION user_service_outbox_pub FOR TABLE outbox_events;
    END IF;
END$$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE outbox_events
    ADD COLUMN topic   VARCHAR(255) NOT NULL DEFAULT 'user.events',
    ADD COLUMN key     VARCHAR(255),
    ADD COLUMN status  VARCHAR(50)  NOT NULL DEFAULT 'pending',
    ADD COLUMN sent_at TIMESTAMPTZ;

UPDATE outbox_events SET key = aggregate_id;
ALTER TABLE outbox_events ALTER COLUMN key SET NOT NULL;

ALTER TABLE outbox_events
    DROP COLUMN aggregate_type,
    DROP COLUMN aggregate_id,
    DROP COLUMN event_type;

DROP INDEX IF EXISTS idx_outbox_events_created_at;
CREATE INDEX idx_outbox_events_status ON outbox_events(status, created_at);

DROP PUBLICATION IF EXISTS user_service_outbox_pub;
