CREATE TABLE IF NOT EXISTS orders (
                                      id BIGSERIAL PRIMARY KEY,
                                      customer_id TEXT NOT NULL,
                                      status TEXT NOT NULL,
                                      total_amount INTEGER NOT NULL,
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                      updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

CREATE TABLE IF NOT EXISTS outbox (
                                      id BIGSERIAL PRIMARY KEY,
                                      aggregate_type TEXT NOT NULL,
                                      aggregate_id BIGINT NOT NULL,
                                      event_type TEXT NOT NULL,
                                      payload JSONB NOT NULL,
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                      published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished
    ON outbox(published_at)
    WHERE published_at IS NULL;

CREATE TABLE IF NOT EXISTS processed_events (
    event_id TEXT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_activity (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    event_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_order_activity_order_id
    ON order_activity(order_id);
