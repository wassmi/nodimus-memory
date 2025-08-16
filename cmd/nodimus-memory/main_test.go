package main

import (
	"errors"
	"os"
	"testing"

	"github.com/wassmi/nodimus-memory/internal/storage"
)

// MockConfig for testing setupCommon
type MockConfig struct {
	ExpandDataDirFunc func() (string, error)
}

func (m *MockConfig) ExpandDataDir() (string, error) {
	if m.ExpandDataDirFunc != nil {
		return m.ExpandDataDirFunc()
	}
	return "", nil
}

// MockDBProvider for testing setupCommon
type MockDBProvider struct {
	NewDBFunc func(dataSourceName string) (*storage.DB, error)
}

func (m *MockDBProvider) NewDB(dataSourceName string) (*storage.DB, error) {
	if m.NewDBFunc != nil {
		return m.NewDBFunc(dataSourceName)
	}
	return nil, errors.New("NewDBFunc not implemented")
}

// MockLogger for testing setupCommon - we no longer need to track Fatalf
type MockLogger struct{}

func (m *MockLogger) Fatalf(format string, v ...interface{}) {}
func (m *MockLogger) Printf(format string, v ...interface{})   {}
func (m *MockLogger) Println(v ...interface{})                {}

func TestSetupCommon(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_data_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("successful setup", func(t *testing.T) {
		mockConfig := &MockConfig{
			ExpandDataDirFunc: func() (string, error) {
				return tmpDir, nil
			},
		}
		mockDBProvider := &MockDBProvider{
			NewDBFunc: func(dataSourceName string) (*storage.DB, error) {
				return storage.NewDB(":memory:")
			},
		}
		mockLogger := &MockLogger{}

		db, dataDir, err := setupCommon(mockLogger, mockConfig, mockDBProvider)

		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if db == nil {
			t.Error("Expected a non-nil DB, got nil")
		}
		if dataDir != tmpDir {
			t.Errorf("Expected dataDir %q, got %q", tmpDir, dataDir)
		}
		if db != nil {
			db.Close()
		}
	})

	t.Run("ExpandDataDir fails", func(t *testing.T) {
		mockConfig := &MockConfig{
			ExpandDataDirFunc: func() (string, error) {
				return "", errors.New("expand error")
			},
		}
		mockDBProvider := &MockDBProvider{}
		mockLogger := &MockLogger{}

		_, _, err := setupCommon(mockLogger, mockConfig, mockDBProvider)

		if err == nil {
			t.Error("Expected an error when ExpandDataDir fails, but got nil")
		}
	})

	t.Run("NewDB fails", func(t *testing.T) {
		mockConfig := &MockConfig{
			ExpandDataDirFunc: func() (string, error) {
				return tmpDir, nil
			},
		}
		mockDBProvider := &MockDBProvider{
			NewDBFunc: func(dataSourceName string) (*storage.DB, error) {
				return nil, errors.New("db open error")
			},
		}
		mockLogger := &MockLogger{}

		_, _, err := setupCommon(mockLogger, mockConfig, mockDBProvider)

		if err == nil {
			t.Error("Expected an error when NewDB fails, but got nil")
		}
	})
}