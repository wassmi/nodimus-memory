package server

import "github.com/wassmi/nodimus-memory/internal/storage"

// DB defines the interface for database operations required by the server.
type DB interface {
	AddMemory(content string, entityNames []string) (int64, error)
	SearchMemories(query string) ([]string, error)
	GetMemory(id int64) (string, error)
	GetEntities() ([]storage.Entity, error)
	GetRelationships() ([]storage.Relationship, error)
}
