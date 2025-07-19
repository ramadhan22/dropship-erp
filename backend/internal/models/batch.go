package models

import "time"

type BatchHistory struct {
	ID          int64      `db:"id" json:"id"`
	ProcessType string     `db:"process_type" json:"process_type"`
	StartedAt   time.Time  `db:"started_at" json:"started_at"`
	EndedAt     *time.Time `db:"ended_at" json:"ended_at"`
	TimeSpent   *string    `db:"time_spent" json:"time_spent"` // PostgreSQL INTERVAL as string
	TotalData   int        `db:"total_data" json:"total_data"`
	DoneData    int        `db:"done_data" json:"done_data"`
	Status      string     `db:"status" json:"status"`
	ErrorMessage string    `db:"error_message" json:"error_message"`
	FileName    string     `db:"file_name" json:"file_name"`
	FilePath    string     `db:"file_path" json:"file_path"`
	CreatedAt   time.Time  `db:"started_at" json:"created_at"` // Use started_at as created_at
	UpdatedAt   time.Time  `db:"started_at" json:"updated_at"` // Placeholder - we could add an actual updated_at column later
}

// BatchHistoryDetail records the result of processing a single transaction within a batch.
type BatchHistoryDetail struct {
	ID        int64  `db:"id" json:"id"`
	BatchID   int64  `db:"batch_id" json:"batch_id"`
	Reference string `db:"reference" json:"reference"`
	Store     string `db:"store" json:"store"`
	Status    string `db:"status" json:"status"`
	ErrorMsg  string `db:"error_message" json:"error_message"`
}
