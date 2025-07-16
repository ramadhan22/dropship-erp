package service

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type JournalRepoInterface interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
	ListJournalEntries(ctx context.Context, from, to, desc string) ([]models.JournalEntry, error)
	GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error)
	GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error)
	ListEntriesBySourceID(ctx context.Context, sourceID string) ([]models.JournalEntry, error)
	DeleteJournalEntry(ctx context.Context, id int64) error
}

type JournalService struct {
	db   *sqlx.DB
	repo JournalRepoInterface
}

func NewJournalService(db *sqlx.DB, r JournalRepoInterface) *JournalService {
	return &JournalService{db: db, repo: r}
}

func (s *JournalService) List(ctx context.Context, from, to, desc string) ([]models.JournalEntry, error) {
	return s.repo.ListJournalEntries(ctx, from, to, desc)
}

func (s *JournalService) Get(ctx context.Context, id int64) (*models.JournalEntry, error) {
	return s.repo.GetJournalEntry(ctx, id)
}

func (s *JournalService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteJournalEntry(ctx, id)
}

func (s *JournalService) Lines(ctx context.Context, id int64) ([]repository.JournalLineDetail, error) {
	return s.repo.GetLinesByJournalID(ctx, id)
}

// EntryWithLines bundles a journal entry with its lines.
type EntryWithLines struct {
	Entry models.JournalEntry            `json:"entry"`
	Lines []repository.JournalLineDetail `json:"lines"`
}

// BulkEntryWithLines bundles a journal entry with model lines for bulk operations
type BulkEntryWithLines struct {
	Entry models.JournalEntry  `json:"entry"`
	Lines []models.JournalLine `json:"lines"`
}

// LinesBySource returns all journal entries matching the source ID along with their lines.
func (s *JournalService) LinesBySource(ctx context.Context, sourceID string) ([]EntryWithLines, error) {
	entries, err := s.repo.ListEntriesBySourceID(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	result := make([]EntryWithLines, 0, len(entries))
	for _, e := range entries {
		lines, err := s.repo.GetLinesByJournalID(ctx, e.JournalID)
		if err != nil {
			return nil, err
		}
		result = append(result, EntryWithLines{Entry: e, Lines: lines})
	}
	return result, nil
}

// Create inserts a JournalEntry along with its lines. It ensures total debits
// equal total credits otherwise returns an error.
func (s *JournalService) Create(
	ctx context.Context,
	e *models.JournalEntry,
	lines []models.JournalLine,
) (int64, error) {
	log.Printf("JournalService.Create %s", e.SourceID)
	var debit, credit float64
	for _, l := range lines {
		if l.IsDebit {
			debit += l.Amount
		} else {
			credit += l.Amount
		}
	}
	if debit != credit {
		logutil.Errorf("JournalService.Create imbalance debit %.2f credit %.2f", debit, credit)
		return 0, fmt.Errorf("debits %.2f do not equal credits %.2f", debit, credit)
	}
	if s.db == nil {
		id, err := s.repo.CreateJournalEntry(ctx, e)
		if err != nil {
			logutil.Errorf("JournalService.Create entry error: %v", err)
			return 0, err
		}
		for i := range lines {
			lines[i].JournalID = id
			if err := s.repo.InsertJournalLine(ctx, &lines[i]); err != nil {
				logutil.Errorf("JournalService.Create line error: %v", err)
				return 0, err
			}
		}
		log.Printf("JournalService.Create done id=%d", id)
		return id, nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		logutil.Errorf("JournalService.Create tx begin error: %v", err)
		return 0, err
	}
	defer tx.Rollback()
	repoTx := repository.NewJournalRepo(tx)
	id, err := repoTx.CreateJournalEntry(ctx, e)
	if err != nil {
		logutil.Errorf("JournalService.Create entry error: %v", err)
		return 0, err
	}
	for i := range lines {
		lines[i].JournalID = id
		if err := repoTx.InsertJournalLine(ctx, &lines[i]); err != nil {
			logutil.Errorf("JournalService.Create line error: %v", err)
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		logutil.Errorf("JournalService.Create commit error: %v", err)
		return 0, err
	}
	log.Printf("JournalService.Create done id=%d", id)
	return id, nil
}

// ========== Performance Optimization Methods ==========

// BulkCreateJournalEntries creates multiple journal entries with their lines in a single transaction
func (s *JournalService) BulkCreateJournalEntries(ctx context.Context, entries []BulkEntryWithLines) ([]int64, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	log.Printf("BulkCreateJournalEntries: processing %d entries", len(entries))

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	repoTx := repository.NewJournalRepo(tx)
	var entryIDs []int64

	for i, entryWithLines := range entries {
		// Validate entry before creating
		if err := s.validateJournalEntry(&entryWithLines.Entry, entryWithLines.Lines); err != nil {
			return nil, fmt.Errorf("validation failed for entry %d: %w", i, err)
		}

		// Create journal entry
		id, err := repoTx.CreateJournalEntry(ctx, &entryWithLines.Entry)
		if err != nil {
			logutil.Errorf("BulkCreateJournalEntries entry %d error: %v", i, err)
			return nil, err
		}

		// Create journal lines
		for j := range entryWithLines.Lines {
			entryWithLines.Lines[j].JournalID = id
			if err := repoTx.InsertJournalLine(ctx, &entryWithLines.Lines[j]); err != nil {
				logutil.Errorf("BulkCreateJournalEntries line %d/%d error: %v", i, j, err)
				return nil, err
			}
		}

		entryIDs = append(entryIDs, id)
	}

	if err := tx.Commit(); err != nil {
		logutil.Errorf("BulkCreateJournalEntries commit error: %v", err)
		return nil, err
	}

	log.Printf("BulkCreateJournalEntries completed: created %d entries", len(entryIDs))
	return entryIDs, nil
}

// validateJournalEntry validates a journal entry and its lines
func (s *JournalService) validateJournalEntry(entry *models.JournalEntry, lines []models.JournalLine) error {
	if entry.Description == nil || *entry.Description == "" {
		return fmt.Errorf("journal entry description cannot be empty")
	}

	if len(lines) == 0 {
		return fmt.Errorf("journal entry must have at least one line")
	}

	// Calculate total debits and credits
	var totalDebits, totalCredits float64
	for _, line := range lines {
		if line.Amount <= 0 {
			return fmt.Errorf("line amount must be positive, got: %f", line.Amount)
		}

		if line.IsDebit {
			totalDebits += line.Amount
		} else {
			totalCredits += line.Amount
		}
	}

	// Check if debits equal credits (within a small tolerance for floating point precision)
	const tolerance = 0.01
	if absFloat(totalDebits-totalCredits) > tolerance {
		return fmt.Errorf("journal entry not balanced: debits=%.2f, credits=%.2f", totalDebits, totalCredits)
	}

	return nil
}

// AutoBalanceJournalEntry automatically balances a journal entry by adjusting the last line
func (s *JournalService) AutoBalanceJournalEntry(entry *models.JournalEntry, lines []models.JournalLine) []models.JournalLine {
	if len(lines) == 0 {
		return lines
	}

	var totalDebits, totalCredits float64
	for i, line := range lines[:len(lines)-1] { // Exclude last line from calculation
		if line.IsDebit {
			totalDebits += line.Amount
		} else {
			totalCredits += line.Amount
		}
		_ = i // Suppress unused variable warning
	}

	// Adjust the last line to balance
	lastLine := &lines[len(lines)-1]
	difference := totalDebits - totalCredits

	if difference > 0 {
		// Need more credits
		lastLine.IsDebit = false
		lastLine.Amount = difference
	} else if difference < 0 {
		// Need more debits
		lastLine.IsDebit = true
		lastLine.Amount = -difference
	} else {
		// Already balanced, keep last line as is
		if lastLine.Amount == 0 {
			lastLine.Amount = 0.01 // Minimum amount
		}
	}

	return lines
}

// BatchDeleteJournalEntries deletes multiple journal entries by source IDs in a single transaction
func (s *JournalService) BatchDeleteJournalEntries(ctx context.Context, sourceIDs []string) error {
	if len(sourceIDs) == 0 {
		return nil
	}

	log.Printf("BatchDeleteJournalEntries: deleting entries for %d sources", len(sourceIDs))

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	repoTx := repository.NewJournalRepo(tx)
	deletedCount := 0

	for _, sourceID := range sourceIDs {
		entries, err := repoTx.ListEntriesBySourceID(ctx, sourceID)
		if err != nil {
			logutil.Errorf("BatchDeleteJournalEntries failed to list entries for source %s: %v", sourceID, err)
			continue
		}

		for _, entry := range entries {
			if err := repoTx.DeleteJournalEntry(ctx, entry.JournalID); err != nil {
				logutil.Errorf("BatchDeleteJournalEntries failed to delete entry %d: %v", entry.JournalID, err)
				continue
			}
			deletedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch delete: %w", err)
	}

	log.Printf("BatchDeleteJournalEntries completed: deleted %d entries", deletedCount)
	return nil
}

// absFloat returns the absolute value of a float64
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
