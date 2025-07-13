package service

import (
	"context"
	"encoding/csv"
	"io"
	"os"
)

// ProcessImportFile reads the CSV file at the given path and stores its rows.
// Batch progress is recorded using the BatchService if available.
func (s *DropshipService) ProcessImportFile(ctx context.Context, batchID int64, path, channel string) {
	f, err := os.Open(path)
	if err != nil {
		if s.batchSvc != nil {
			s.batchSvc.UpdateStatus(ctx, batchID, "failed", err.Error())
		}
		return
	}
	defer f.Close()

	var total int
	if s.batchSvc != nil {
		total, err = countCSVRows(f)
		if err != nil {
			s.batchSvc.UpdateStatus(ctx, batchID, "failed", err.Error())
			return
		}
		s.batchSvc.UpdateTotal(ctx, batchID, total)
		s.batchSvc.UpdateStatus(ctx, batchID, "processing", "")
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			s.batchSvc.UpdateStatus(ctx, batchID, "failed", err.Error())
			return
		}
	}

	count, err := s.ImportFromCSV(ctx, f, channel, batchID)
	if s.batchSvc != nil {
		s.batchSvc.UpdateDone(ctx, batchID, count)
		if err != nil {
			s.batchSvc.UpdateStatus(ctx, batchID, "failed", err.Error())
		} else {
			s.batchSvc.UpdateStatus(ctx, batchID, "completed", "")
		}
	}
}

// countCSVRows returns the number of data rows in a CSV reader.
// It expects the first line to be a header and ignores it.
func countCSVRows(r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	if _, err := reader.Read(); err != nil {
		return 0, err
	}
	n := 0
	for {
		if _, err := reader.Read(); err == io.EOF {
			break
		} else if err != nil {
			return n, err
		} else {
			n++
		}
	}
	return n, nil
}
