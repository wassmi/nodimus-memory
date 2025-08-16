package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/nodimus/nodimus/internal/kg"
)

// Server is the JSON-RPC 2.0 server.
type Server struct {
	*http.Server
}

// NewServer creates a new JSON-RPC 2.0 server.
func NewServer(port int, bind string, timeout int, service *MemoryService) *Server {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(json2.NewCodec(), "application/json")
	rpcServer.RegisterService(service, "memory")

	router := mux.NewRouter()
	router.Handle("/rpc", rpcServer)

	return &Server{
		Server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", bind, port),
			Handler:      router,
			ReadTimeout:  time.Duration(timeout) * time.Second,
			WriteTimeout: time.Duration(timeout) * time.Second,
		},
	}
}

// Start starts the server.
func (s *Server) Start() error {
	return s.ListenAndServe()
}

// Stop stops the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

// MemoryService is the service that provides the MCP capabilities.
type MemoryService struct {
	DB      DB
	DataDir string
	Log     *log.Logger
}

// AddMemoryRequest is the request for the AddMemory method.
type AddMemoryRequest struct {
	Content  string   `json:"content"`
	Entities []string `json:"entities"`
}

// AddMemoryResponse is the response for the AddMemory method.
type AddMemoryResponse struct {
	ID int64 `json:"id"`
}

// AddMemory adds a new memory to the database.
func (s *MemoryService) AddMemory(r *http.Request, args *AddMemoryRequest, reply *AddMemoryResponse) error {
	id, err := s.DB.AddMemory(args.Content, args.Entities)
	if err != nil {
		return err
	}
	reply.ID = id

	// Regenerate the knowledge graph in the background.
	go func() {
		if err := kg.Generate(s.DB, filepath.Join(s.DataDir, "knowledge-graph.jsonld")); err != nil {
			s.Log.Printf("failed to regenerate knowledge graph: %v\n", err)
		}
	}()

	return nil
}

// SearchMemoryRequest is the request for the SearchMemory method.
type SearchMemoryRequest struct {
	Query string `json:"query"`
}

// SearchMemoryResponse is the response for the SearchMemory method.
type SearchMemoryResponse struct {
	Results []string `json:"results"`
}

// SearchMemory searches for memories in the database.
func (s *MemoryService) SearchMemory(r *http.Request, args *SearchMemoryRequest, reply *SearchMemoryResponse) error {
	results, err := s.DB.SearchMemories(args.Query)
	if err != nil {
		return err
	}
	reply.Results = results
	return nil
}

// GetContextRequest is the request for the GetContext method.
type GetContextRequest struct {
	ID int64 `json:"id"`
}

// GetContextResponse is the response for the GetContext method.
type GetContextResponse struct {
	Context string `json:"context"`
}

// GetContext gets the context for a given memory.
func (s *MemoryService) GetContext(r *http.Request, args *GetContextRequest, reply *GetContextResponse) error {
	context, err := s.DB.GetMemory(args.ID)
	if err != nil {
		return err
	}
	reply.Context = context
	return nil
}
