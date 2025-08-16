package snapshot

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/wassmi/nodimus-memory/internal/storage"
)

// SnapshotDB defines the database operations required by the snapshot package.
type SnapshotDB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Snapshotter is a database snapshotter.
type Snapshotter struct {
	cron *cron.Cron
}

// NewSnapshotter creates a new snapshotter.
func NewSnapshotter() *Snapshotter {
	return &Snapshotter{
		cron: cron.New(),
	}
}

// Start starts the snapshotter.
func (s *Snapshotter) Start(db *storage.DB, dataDir string) error {
	_, err := s.cron.AddFunc("@daily", func() {
		snapshotDir := filepath.Join(dataDir, "snapshots")
		if err := os.MkdirAll(snapshotDir, 0755); err != nil {
			fmt.Printf("failed to create snapshot directory: %v\n", err)
			return
		}

		date := time.Now().Format("2006-01-02")
		snapshotFile := filepath.Join(snapshotDir, fmt.Sprintf("%s.db", date))

		_, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", snapshotFile))
		if err != nil {
			fmt.Printf("failed to create snapshot: %v\n", err)
			return
		}

		fmt.Printf("created snapshot: %s\n", snapshotFile)
	})
	if err != nil {
		return err
	}

	s.cron.Start()

	return nil
}

// Stop stops the snapshotter.
func (s *Snapshotter) Stop() {
	s.cron.Stop()
}

// IntegrityCheck checks the integrity of the database.
func IntegrityCheck(db SnapshotDB) error {
	_, err := db.Exec("PRAGMA integrity_check")
	return err
}