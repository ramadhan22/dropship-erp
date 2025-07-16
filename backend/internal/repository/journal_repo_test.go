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

func TestInsertJournalLinesBulk(t *testing.T) {
	ctx := context.Background()
	jrepo := NewJournalRepo(testDB)
	now := time.Now()

	// 1. Insert test accounts
	acc1 := insertTestAccount(t, testDB, "BULK100", "Bulk Test Asset", "Asset", nil)
	acc2 := insertTestAccount(t, testDB, "BULK200", "Bulk Test Expense", "Expense", nil)
	acc3 := insertTestAccount(t, testDB, "BULK300", "Bulk Test Revenue", "Revenue", nil)

	// 2. Create a JournalEntry
	je := &models.JournalEntry{
		EntryDate:   now,
		Description: ptrString("Bulk Insert Test"),
		SourceType:  "bulk_test", SourceID: "BULK-1", ShopUsername: "BulkShop",
	}
	journalID, err := jrepo.CreateJournalEntry(ctx, je)
	if err != nil {
		t.Fatalf("CreateJournalEntry failed: %v", err)
	}

	// 3. Test bulk insert with multiple lines
	lines := []models.JournalLine{
		{JournalID: journalID, AccountID: acc1, IsDebit: true, Amount: 150, Memo: ptrString("Bulk debit 1")},
		{JournalID: journalID, AccountID: acc2, IsDebit: true, Amount: 50, Memo: ptrString("Bulk debit 2")},
		{JournalID: journalID, AccountID: acc3, IsDebit: false, Amount: 200, Memo: ptrString("Bulk credit")},
	}

	if err := jrepo.InsertJournalLines(ctx, lines); err != nil {
		t.Fatalf("InsertJournalLines failed: %v", err)
	}

	// 4. Verify all lines were inserted
	insertedLines, err := jrepo.GetLinesByJournalID(ctx, journalID)
	if err != nil {
		t.Fatalf("GetLinesByJournalID failed: %v", err)
	}

	if len(insertedLines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(insertedLines))
	}

	// 5. Test edge cases

	// Test empty slice
	if err := jrepo.InsertJournalLines(ctx, []models.JournalLine{}); err != nil {
		t.Errorf("InsertJournalLines with empty slice should not error: %v", err)
	}

	// Test single line (should use InsertJournalLine internally)
	singleLine := []models.JournalLine{
		{JournalID: journalID, AccountID: acc1, IsDebit: true, Amount: 25, Memo: ptrString("Single line test")},
	}
	if err := jrepo.InsertJournalLines(ctx, singleLine); err != nil {
		t.Fatalf("InsertJournalLines with single line failed: %v", err)
	}

	// Verify the single line was inserted
	finalLines, err := jrepo.GetLinesByJournalID(ctx, journalID)
	if err != nil {
		t.Fatalf("GetLinesByJournalID failed: %v", err)
	}

	if len(finalLines) != 4 {
		t.Errorf("Expected 4 lines total, got %d", len(finalLines))
	}

	// 6. Cleanup
	testDB.ExecContext(ctx, "DELETE FROM journal_entries WHERE journal_id = $1", journalID)
	tempCleanupAccount(t, testDB, acc1)
	tempCleanupAccount(t, testDB, acc2)
	tempCleanupAccount(t, testDB, acc3)
}

func TestInsertJournalLinesValidation(t *testing.T) {
	ctx := context.Background()
	jrepo := NewJournalRepo(testDB)
	now := time.Now()

	// 1. Insert test accounts
	acc1 := insertTestAccount(t, testDB, "VAL100", "Validation Test Asset", "Asset", nil)
	acc2 := insertTestAccount(t, testDB, "VAL200", "Validation Test Expense", "Expense", nil)

	// 2. Create a JournalEntry
	je := &models.JournalEntry{
		EntryDate:   now,
		Description: ptrString("Validation Test"),
		SourceType:  "validation_test", SourceID: "VAL-1", ShopUsername: "ValidationShop",
	}
	journalID, err := jrepo.CreateJournalEntry(ctx, je)
	if err != nil {
		t.Fatalf("CreateJournalEntry failed: %v", err)
	}

	// 3. Test balanced lines (should succeed)
	balancedLines := []models.JournalLine{
		{JournalID: journalID, AccountID: acc1, IsDebit: true, Amount: 100, Memo: ptrString("Balanced debit")},
		{JournalID: journalID, AccountID: acc2, IsDebit: false, Amount: 100, Memo: ptrString("Balanced credit")},
	}

	if err := jrepo.InsertJournalLines(ctx, balancedLines); err != nil {
		t.Errorf("InsertJournalLines with balanced amounts should succeed: %v", err)
	}

	// 4. Test unbalanced lines (should fail)
	unbalancedLines := []models.JournalLine{
		{JournalID: journalID, AccountID: acc1, IsDebit: true, Amount: 150, Memo: ptrString("Unbalanced debit")},
		{JournalID: journalID, AccountID: acc2, IsDebit: false, Amount: 100, Memo: ptrString("Unbalanced credit")},
	}

	err = jrepo.InsertJournalLines(ctx, unbalancedLines)
	if err == nil {
		t.Error("InsertJournalLines with unbalanced amounts should fail")
	}
	expectedErrorMsg := "debits 150.00 do not equal credits 100.00"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// 5. Test complex balanced scenario with multiple debits and credits
	complexBalancedLines := []models.JournalLine{
		{JournalID: journalID, AccountID: acc1, IsDebit: true, Amount: 75, Memo: ptrString("Debit 1")},
		{JournalID: journalID, AccountID: acc1, IsDebit: true, Amount: 25, Memo: ptrString("Debit 2")},
		{JournalID: journalID, AccountID: acc2, IsDebit: false, Amount: 60, Memo: ptrString("Credit 1")},
		{JournalID: journalID, AccountID: acc2, IsDebit: false, Amount: 40, Memo: ptrString("Credit 2")},
	}

	if err := jrepo.InsertJournalLines(ctx, complexBalancedLines); err != nil {
		t.Errorf("InsertJournalLines with complex balanced amounts should succeed: %v", err)
	}

	// 6. Cleanup
	testDB.ExecContext(ctx, "DELETE FROM journal_entries WHERE journal_id = $1", journalID)
	tempCleanupAccount(t, testDB, acc1)
	tempCleanupAccount(t, testDB, acc2)
}
