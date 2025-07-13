CREATE UNIQUE INDEX IF NOT EXISTS journal_entries_source_idx
    ON journal_entries (source_type, source_id);
