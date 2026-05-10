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
