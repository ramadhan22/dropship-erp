package logutil

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DailyFileWriter writes logs to a file named by today's date inside dir.
// It automatically rotates when the day changes.
type DailyFileWriter struct {
	dir  string
	date string
	mu   sync.Mutex
	f    *os.File
}

// NewDailyFileWriter creates a writer that stores logs in dir/YYYY-MM-DD.log.
func NewDailyFileWriter(dir string) (*DailyFileWriter, error) {
	w := &DailyFileWriter{dir: dir}
	return w, w.rotate()
}

func (w *DailyFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.rotate(); err != nil {
		return 0, err
	}
	return w.f.Write(p)
}

func (w *DailyFileWriter) rotate() error {
	today := time.Now().Format("2006-01-02")
	if w.date == today && w.f != nil {
		return nil
	}
	if w.f != nil {
		w.f.Close()
	}
	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(w.dir, today+".log")
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.f = f
	w.date = today
	return nil
}

// Close closes the underlying file.
func (w *DailyFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f != nil {
		return w.f.Close()
	}
	return nil
}

var _ io.WriteCloser = (*DailyFileWriter)(nil)
