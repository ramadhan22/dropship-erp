// File: backend/internal/repository/journal_repo_test.go

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// insertTestAccount inserts a dummy account and returns its account_id.
func insertTestAccount(t *testing.T, db *sqlx.DB, code, name, aType string, parentID *int64) int64 {
	var id int64
	err := db.QueryRowxContext(context.Background(),
		`INSERT INTO accounts (account_code, account_name, account_type, parent_id)
         VALUES ($1, $2, $3, $4) RETURNING account_id`,
		code, name, aType, parentID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertTestAccount failed: %v", err)
	}
	return id
}

// tempCleanupAccount deletes the account by its ID.
func tempCleanupAccount(t *testing.T, db *sqlx.DB, accountID int64) {
	_, err := db.ExecContext(context.Background(),
		"DELETE FROM accounts WHERE account_id = $1", accountID)
	if err != nil {
		t.Fatalf("cleanup account failed: %v", err)
	}
}

func TestCreateJournalEntryAndLinesAndBalance(t *testing.T) {
	ctx := context.Background()
	jrepo := NewJournalRepo(testDB)
	now := time.Now()

	// 1. Insert two test accounts: Asset and Expense
	acc1 := insertTestAccount(t, testDB, "TEST100", "Test Asset", "Asset", nil)
	acc2 := insertTestAccount(t, testDB, "TEST200", "Test Expense", "Expense", nil)
	t.Logf("Inserted accounts %d and %d", acc1, acc2)

	// 2. Create a JournalEntry
	je := &models.JournalEntry{
		EntryDate:   now,
		Description: ptrString("Test Journal Entry"),
		SourceType:  "test", SourceID: "SRC-1", ShopUsername: "TestShop",
	}
	journalID, err := jrepo.CreateJournalEntry(ctx, je)
	if err != nil {
		t.Fatalf("CreateJournalEntry failed: %v", err)
	}
	t.Logf("Created JournalEntry ID %d", journalID)

	// 3. Insert debit (acc1) and credit (acc2) lines
	jl1 := &models.JournalLine{
		JournalID: journalID,
		AccountID: acc1,
		IsDebit:   true,
		Amount:    100,
		Memo:      ptrString("Debit test"),
	}
	if err := jrepo.InsertJournalLine(ctx, jl1); err != nil {
		t.Fatalf("InsertJournalLine 1 failed: %v", err)
	}
	jl2 := &models.JournalLine{
		JournalID: journalID,
		AccountID: acc2,
		IsDebit:   false,
		Amount:    100,
		Memo:      ptrString("Credit test"),
	}
	if err := jrepo.InsertJournalLine(ctx, jl2); err != nil {
		t.Fatalf("InsertJournalLine 2 failed: %v", err)
	}
	t.Log("Inserted JournalLines")

	// 4. Get lines by shop and date range
	lines, err := jrepo.GetJournalLinesByShopAndDate(
		ctx, "TestShop", now.Add(-time.Hour), now.Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("GetJournalLinesByShopAndDate failed: %v", err)
	}
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 rows, got %d", len(lines))
	}

	// 5. Check account balances as of now
	balances, err := jrepo.GetAccountBalancesAsOf(ctx, "TestShop", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("GetAccountBalancesAsOf failed: %v", err)
	}
	// Expect acc1 balance = +100, acc2 = -100
	foundAcc1, foundAcc2 := false, false
	for _, ab := range balances {
		if ab.AccountID == acc1 {
			foundAcc1 = true
			if ab.Balance != 100 {
				t.Errorf("Expected balance 100 for acc1, got %f", ab.Balance)
			}
		}
		if ab.AccountID == acc2 {
			foundAcc2 = true
			if ab.Balance != -100 {
				t.Errorf("Expected balance -100 for acc2, got %f", ab.Balance)
			}
		}
	}
	if !foundAcc1 || !foundAcc2 {
		t.Errorf("Did not find balances for both test accounts: %v, %v", foundAcc1, foundAcc2)
	}

	// 6. Check balances within date range
	rangeBalances, err := jrepo.GetAccountBalancesBetween(ctx, "TestShop", now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("GetAccountBalancesBetween failed: %v", err)
	}
	if len(rangeBalances) == 0 {
		t.Fatalf("Expected balances, got 0")
	}

	foundAcc1, foundAcc2 = false, false
	for _, ab := range rangeBalances {
		if ab.AccountID == acc1 {
			foundAcc1 = true
			if ab.Balance != 100 {
				t.Errorf("Expected range balance 100 for acc1, got %f", ab.Balance)
			}
		}
		if ab.AccountID == acc2 {
			foundAcc2 = true
			if ab.Balance != -100 {
				t.Errorf("Expected range balance -100 for acc2, got %f", ab.Balance)
			}
		}
	}
	if !foundAcc1 || !foundAcc2 {
		t.Errorf("Did not find range balances for both test accounts: %v, %v", foundAcc1, foundAcc2)
	}
	// 7. Cleanup: delete journal entry and the two accounts
	testDB.ExecContext(ctx, "DELETE FROM journal_entries WHERE journal_id = $1", journalID)
	tempCleanupAccount(t, testDB, acc1)
	tempCleanupAccount(t, testDB, acc2)
}
