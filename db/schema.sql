CREATE TABLE IF NOT EXISTS payment_attempts (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    provider_name TEXT NOT NULL,
    provider_payment_id TEXT,
    status TEXT NOT NULL,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    return_url TEXT NOT NULL,
    failure_reason TEXT NOT NULL DEFAULT '',
    client_secret TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS processed_webhook_events (
    event_id TEXT PRIMARY KEY,
    provider_name TEXT NOT NULL,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS payment_attempts_idempotency_key_uidx
    ON payment_attempts (idempotency_key);

CREATE UNIQUE INDEX IF NOT EXISTS payment_attempts_provider_name_provider_payment_id_uidx
    ON payment_attempts (provider_name, provider_payment_id)
    WHERE provider_payment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS payment_attempts_order_id_idx
    ON payment_attempts (order_id);

CREATE INDEX IF NOT EXISTS payment_attempts_status_idx
    ON payment_attempts (status);

CREATE INDEX IF NOT EXISTS processed_webhook_events_provider_name_idx
    ON processed_webhook_events (provider_name);

CREATE INDEX IF NOT EXISTS processed_webhook_events_processed_at_idx
    ON processed_webhook_events (processed_at);
