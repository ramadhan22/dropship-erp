package models

type BatchHistory struct {
	ID          int64  `db:"id" json:"id"`
	ProcessType string `db:"process_type" json:"process_type"`
	StartedAt   string `db:"started_at" json:"started_at"`
	TotalData   int    `db:"total_data" json:"total_data"`
	DoneData    int    `db:"done_data" json:"done_data"`
	Status      string `db:"status" json:"status"`
	ErrorMsg    string `db:"error_message" json:"error_message"`
	FileName    string `db:"file_name" json:"file_name"`
	FilePath    string `db:"file_path" json:"file_path"`
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
