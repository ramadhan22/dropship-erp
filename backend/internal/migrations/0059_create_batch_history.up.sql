CREATE TABLE batch_history (
    id SERIAL PRIMARY KEY,
    process_type TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    total_data INTEGER NOT NULL,
    done_data INTEGER NOT NULL DEFAULT 0
);
