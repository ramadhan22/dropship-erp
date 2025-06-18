package models

import "time"

type AssetAccount struct {
	ID        int64     `db:"id" json:"id"`
	AccountID int64     `db:"account_id" json:"account_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
