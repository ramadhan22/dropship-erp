package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestBatchRepo_UpdateStatusWithEndTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewBatchRepo(sqlxDB)

	// Mock the UPDATE query for UpdateStatusWithEndTime
	mock.ExpectExec(`UPDATE batch_history`).
		WithArgs(int64(1), "completed", "").
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.UpdateStatusWithEndTime(ctx, 1, "completed", "")
	
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBatchRepo_Insert_WithNewFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewBatchRepo(sqlxDB)

	// Mock the INSERT query
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(123))
	mock.ExpectQuery(`INSERT INTO batch_history`).
		WillReturnRows(rows)

	ctx := context.Background()
	batch := &models.BatchHistory{
		ProcessType:  "test_process",
		TotalData:    10,
		DoneData:     0,
		Status:       "pending",
		ErrorMessage: "",
		FileName:     "test.csv",
		FilePath:     "/tmp/test.csv",
		EndedAt:      nil,
		TimeSpent:    nil,
	}

	id, err := repo.Insert(ctx, batch)
	
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}