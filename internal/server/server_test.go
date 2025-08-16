package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wassmi/nodimus-memory/internal/storage"
)

// MockDB implements the DB interface for testing.
type MockDB struct {
	AddMemoryFunc        func(content string, entityNames []string) (int64, error)
	SearchMemoriesFunc   func(query string) ([]string, error)
	GetMemoryFunc        func(id int64) (string, error)
	GetEntitiesFunc      func() ([]storage.Entity, error)
	GetRelationshipsFunc func() ([]storage.Relationship, error)
}

func (m *MockDB) AddMemory(content string, entityNames []string) (int64, error) {
	return m.AddMemoryFunc(content, entityNames)
}
func (m *MockDB) SearchMemories(query string) ([]string, error) {
	return m.SearchMemoriesFunc(query)
}
func (m *MockDB) GetMemory(id int64) (string, error) {
	return m.GetMemoryFunc(id)
}
func (m *MockDB) GetEntities() ([]storage.Entity, error) {
	return m.GetEntitiesFunc()
}
func (m *MockDB) GetRelationships() ([]storage.Relationship, error) {
	return m.GetRelationshipsFunc()
}

func TestAddMemory(t *testing.T) {
	mockDB := &MockDB{
		AddMemoryFunc: func(content string, entityNames []string) (int64, error) {
			if content == "test content" && len(entityNames) == 1 && entityNames[0] == "test entity" {
				return 1, nil
			}
			return 0, errors.New("invalid input")
		},
		GetEntitiesFunc:      func() ([]storage.Entity, error) { return nil, nil },
		GetRelationshipsFunc: func() ([]storage.Relationship, error) { return nil, nil },
	}

	// Create a temporary directory for the knowledge graph
	tmpDir, err := os.MkdirTemp("", "test_kg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock logger
	var logBuffer bytes.Buffer
	mockLogger := log.New(&logBuffer, "", 0)

	service := &MemoryService{
		DB:      mockDB,
		DataDir: tmpDir,
		Log:     mockLogger,
	}

	req := &AddMemoryRequest{
		Content:  "test content",
		Entities: []string{"test entity"},
	}
	reply := &AddMemoryResponse{}

	err = service.AddMemory(nil, req, reply)
	if err != nil {
		t.Fatalf("AddMemory failed: %v", err)
	}

	if reply.ID != 1 {
		t.Errorf("Expected ID 1, got %d", reply.ID)
	}

	// Give some time for the goroutine to generate the KG
	time.Sleep(100 * time.Millisecond)

	// Check if the knowledge graph file was created
	kgPath := filepath.Join(tmpDir, "knowledge-graph.jsonld")
	if _, err := os.Stat(kgPath); os.IsNotExist(err) {
		t.Errorf("Knowledge graph file %s was not created", kgPath)
	}
}

func TestSearchMemory(t *testing.T) {
	mockDB := &MockDB{
		SearchMemoriesFunc: func(query string) ([]string, error) {
			if query == "test query" {
				return []string{"memory 1", "memory 2"}, nil
			}
			return nil, errors.New("no results")
		},
	}

	service := &MemoryService{
		DB: mockDB,
	}

	req := &SearchMemoryRequest{
		Query: "test query",
	}
	reply := &SearchMemoryResponse{}

	err := service.SearchMemory(nil, req, reply)
	if err != nil {
		t.Fatalf("SearchMemory failed: %v", err)
	}

	if len(reply.Results) != 2 || reply.Results[0] != "memory 1" || reply.Results[1] != "memory 2" {
		t.Errorf("Expected [\"memory 1\", \"memory 2\"], got %v", reply.Results)
	}
}

func TestGetContext(t *testing.T) {
	mockDB := &MockDB{
		GetMemoryFunc: func(id int64) (string, error) {
			if id == 1 {
				return "context for id 1", nil
			}
			return "", errors.New("not found")
		},
	}

	service := &MemoryService{
		DB: mockDB,
	}

	req := &GetContextRequest{
		ID: 1,
	}
	reply := &GetContextResponse{}

	err := service.GetContext(nil, req, reply)
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}

	if reply.Context != "context for id 1" {
		t.Errorf("Expected \"context for id 1\", got %s", reply.Context)
	}
}

func TestNewServer(t *testing.T) {
	// Create a mock service
	mockService := &MemoryService{
		DB: &MockDB{ // Provide a mock DB that satisfies all methods
			AddMemoryFunc:        func(content string, entityNames []string) (int64, error) { return 0, nil },
			SearchMemoriesFunc:   func(query string) ([]string, error) { return nil, nil },
			GetMemoryFunc:        func(id int64) (string, error) { return "", nil },
			GetEntitiesFunc:      func() ([]storage.Entity, error) { return nil, nil },
			GetRelationshipsFunc: func() ([]storage.Relationship, error) { return nil, nil },
		},
		DataDir: "/tmp",
		Log:     log.New(os.Stdout, "", 0),
	}

	server := NewServer(8080, "127.0.0.1", 1, mockService)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	// Test server start and stop (non-blocking)
	go func() {
		err := server.Start()
		if err != http.ErrServerClosed {
			t.Errorf("Server Start returned unexpected error: %v", err)
		}
	}()

	// Give server some time to start
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Fatalf("Server Stop failed: %v", err)
	}
}

func TestServerRPCMethods(t *testing.T) {
	mockDB := &MockDB{
		AddMemoryFunc: func(content string, entityNames []string) (int64, error) {
			return 123, nil
		},
		SearchMemoriesFunc: func(query string) ([]string, error) {
			return []string{"found memory"}, nil
		},
		GetMemoryFunc: func(id int64) (string, error) {
			return "retrieved context", nil
		},
		GetEntitiesFunc:      func() ([]storage.Entity, error) { return nil, nil },
		GetRelationshipsFunc: func() ([]storage.Relationship, error) { return nil, nil },
	}

	// Create a temporary directory for the knowledge graph
	tmpDir, err := os.MkdirTemp("", "test_kg_rpc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock logger
	var logBuffer bytes.Buffer
	mockLogger := log.New(&logBuffer, "", 0)

	service := &MemoryService{
		DB:      mockDB,
		DataDir: tmpDir,
		Log:     mockLogger,
	}

	// Create a test server
	ts := httptest.NewServer(NewServer(0, "127.0.0.1", 1, service).Handler)
	defer ts.Close()

	// Test AddMemory via RPC
	addReq := `{"jsonrpc":"2.0","method":"memory.AddMemory","params":[{"content":"rpc test","entities":["rpc"]}],"id":1}`
	resp, err := http.Post(ts.URL+"/rpc", "application/json", strings.NewReader(addReq))
	if err != nil {
		t.Fatalf("RPC AddMemory request failed: %v", err)
	}
	defer resp.Body.Close()

	var addReply struct {
		Jsonrpc string `json:"jsonrpc"`
		Result  struct {
			ID int64 `json:"id"`
		} `json:"result"`
		ID      int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&addReply); err != nil {
		t.Fatalf("Failed to decode AddMemory response: %v", err)
	}
	if addReply.Result.ID != 123 {
		t.Errorf("Expected AddMemory ID 123, got %d", addReply.Result.ID)
	}

	// Test SearchMemory via RPC
	searchReq := `{"jsonrpc":"2.0","method":"memory.SearchMemory","params":[{"query":"rpc test"}],"id":1}`
	resp, err = http.Post(ts.URL+"/rpc", "application/json", strings.NewReader(searchReq))
	if err != nil {
		t.Fatalf("RPC SearchMemory request failed: %v", err)
	}
	defer resp.Body.Close()

	var searchReply struct {
		Jsonrpc string `json:"jsonrpc"`
		Result  struct {
			Results []string `json:"results"`
		} `json:"result"`
		ID      int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchReply); err != nil {
		t.Fatalf("Failed to decode SearchMemory response: %v", err)
	}
	if len(searchReply.Result.Results) != 1 || searchReply.Result.Results[0] != "found memory" {
		t.Errorf("Expected SearchMemory results [\"found memory\"], got %v", searchReply.Result.Results)
	}

	// Test GetContext via RPC
	getContextReq := `{"jsonrpc":"2.0","method":"memory.GetContext","params":[{"id":1}],"id":1}`
	resp, err = http.Post(ts.URL+"/rpc", "application/json", strings.NewReader(getContextReq))
	if err != nil {
		t.Fatalf("RPC GetContext request failed: %v", err)
	}
	defer resp.Body.Close()

	var getContextReply struct {
		Jsonrpc string `json:"jsonrpc"`
		Result  struct {
			Context string `json:"context"`
		} `json:"result"`
		ID      int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&getContextReply); err != nil {
		t.Fatalf("Failed to decode GetContext response: %v", err)
	}
	if getContextReply.Result.Context != "retrieved context" {
		t.Errorf("Expected GetContext \"retrieved context\", got %s", getContextReply.Result.Context)
	}
}
