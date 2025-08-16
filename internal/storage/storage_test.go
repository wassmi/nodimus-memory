package storage

import (
	"testing"
)

func TestDB(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	content := "test memory"
	id, err := db.AddMemory(content, []string{})
	if err != nil {
		t.Fatalf("failed to add memory: %v", err)
	}

	if id != 1 {
		t.Errorf("expected memory id to be 1, got %d", id)
	}

	memories, err := db.SearchMemories("test")
	if err != nil {
		t.Fatalf("failed to search memories: %v", err)
	}

	if len(memories) != 1 {
		t.Errorf("expected to find 1 memory, got %d", len(memories))
	}

	if memories[0] != content {
		t.Errorf("expected memory content to be '%s', got '%s'", content, memories[0])
	}
}