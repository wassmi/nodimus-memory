package kg

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/wassmi/nodimus-memory/internal/storage"
)

// MockDB implements the KGDB interface for testing.
type MockDB struct {
	GetEntitiesFunc     func() ([]storage.Entity, error)
	GetRelationshipsFunc func() ([]storage.Relationship, error)
}

func (m *MockDB) GetEntities() ([]storage.Entity, error) {
	return m.GetEntitiesFunc()
}

func (m *MockDB) GetRelationships() ([]storage.Relationship, error) {
	return m.GetRelationshipsFunc()
}

func TestGenerate(t *testing.T) {
	// Create a temporary file for the knowledge graph
	kgFile := "test_kg.jsonld"
	defer os.Remove(kgFile)

	// Mock data
	mockEntities := []storage.Entity{
		{ID: 1, Name: "Paris", Type: "City"},
		{ID: 2, Name: "France", Type: "Country"},
	}
	mockRelationships := []storage.Relationship{
		{ID: 101, SourceID: 1, TargetID: 2, Type: "LOCATED_IN"},
	}

	// Create a mock DB
	mockDB := &MockDB{
		GetEntitiesFunc: func() ([]storage.Entity, error) {
			return mockEntities, nil
		},
		GetRelationshipsFunc: func() ([]storage.Relationship, error) {
			return mockRelationships, nil
		},
	}

	// Generate the knowledge graph
	err := Generate(mockDB, kgFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Read and unmarshal the generated file
	content, err := ioutil.ReadFile(kgFile)
	if err != nil {
		t.Fatalf("Failed to read generated KG file: %v", err)
	}

	var graph map[string]interface{}
	if err := json.Unmarshal(content, &graph); err != nil {
		t.Fatalf("Failed to unmarshal generated KG: %v", err)
	}

	// Assertions
	if context, ok := graph["@context"]; !ok || context != "https://schema.org/" {
		t.Errorf("Expected @context to be https://schema.org/, got %v", context)
	}

	graphNodes, ok := graph["@graph"].([]interface{})
	if !ok || len(graphNodes) != 3 { // 2 entities + 1 relationship
		t.Errorf("Expected 3 graph nodes, got %d", len(graphNodes))
	}

	// Further assertions can be made to check the content of each node
}