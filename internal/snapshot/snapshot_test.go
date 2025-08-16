package snapshot

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/wassmi/nodimus-memory/internal/storage"
)

// MockDB for snapshot tests
type MockDB struct {
	ExecFunc func(query string, args ...interface{}) (sql.Result, error)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.ExecFunc(query, args...)
}

func TestIntegrityCheck(t *testing.T) {
	mockDB := &MockDB{
		ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
			if query == "PRAGMA integrity_check" {
				return nil, nil // Simulate success
			}
			return nil, errors.New("unexpected query")
		},
	}

	err := IntegrityCheck(mockDB)
	if err != nil {
		t.Errorf("IntegrityCheck failed: %v", err)
	}

	// Test with an error
	mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
		return nil, errors.New("integrity check failed")
	}
	err = IntegrityCheck(mockDB)
	if err == nil {
		t.Error("IntegrityCheck did not return an error when expected")
	}
}
