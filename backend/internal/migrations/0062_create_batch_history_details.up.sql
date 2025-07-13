CREATE TABLE IF NOT EXISTS batch_history_details (
    id SERIAL PRIMARY KEY,
    batch_id INTEGER NOT NULL REFERENCES batch_history(id) ON DELETE CASCADE,
    reference TEXT NOT NULL,
    store TEXT,
    status TEXT NOT NULL,
    error_message TEXT
);
