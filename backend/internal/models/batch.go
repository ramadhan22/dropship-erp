package models

type BatchHistory struct {
	ID          int64  `db:"id" json:"id"`
	ProcessType string `db:"process_type" json:"process_type"`
	StartedAt   string `db:"started_at" json:"started_at"`
	TotalData   int    `db:"total_data" json:"total_data"`
	DoneData    int    `db:"done_data" json:"done_data"`
}
