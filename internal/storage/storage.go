package storage

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"strconv"

	"github.com/blevesearch/bleve/v2"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// DB is a wrapper around the SQL database connection.
type DB struct {
	*sql.DB
	index bleve.Index
}

// NewDB creates a new database connection.
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Open or create a bleve index
	indexPath := dataSourceName + ".bleve"
	var index bleve.Index
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Index does not exist, create it
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create bleve index: %w", err)
		}
	} else {
		// Index exists, open it
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open bleve index: %w", err)
		}
	}

	return &DB{DB: db, index: index}, nil
}

// Migrate runs the database migrations.
func (db *DB) Migrate() error {
	_, err := db.Exec(schema)
	return err
}

// AddMemory adds a new memory and links it to the given entities.
func (db *DB) AddMemory(content string, entityNames []string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec("INSERT INTO memories (content) VALUES (?)", content)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	memoryID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, entityName := range entityNames {
		var entityID int64
		err := tx.QueryRow("SELECT id FROM entities WHERE name = ?", entityName).Scan(&entityID)
		if err == sql.ErrNoRows {
			result, err := tx.Exec("INSERT INTO entities (name, type) VALUES (?, ?)", entityName, "unknown")
			if err != nil {
				tx.Rollback()
				return 0, err
			}
			entityID, err = result.LastInsertId()
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		} else if err != nil {
			tx.Rollback()
			return 0, err
		}

		_, err = tx.Exec("INSERT INTO memory_entities (memory_id, entity_id) VALUES (?, ?)", memoryID, entityID)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if err := db.index.Index(strconv.FormatInt(memoryID, 10), content); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to index memory: %w", err)
	}

	return memoryID, tx.Commit()
}

// SearchMemories searches for memories in the bleve index.
func (db *DB) SearchMemories(query string) ([]string, error) {
	queryObj := bleve.NewMatchQuery(query)
	searchRequest := bleve.NewSearchRequest(queryObj)
	searchResult, err := db.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search index: %w", err)
	}

	var memories []string
	for _, hit := range searchResult.Hits {
		id, err := strconv.ParseInt(hit.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse memory ID: %w", err)
		}
		content, err := db.GetMemory(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get memory content: %w", err)
		}
		memories = append(memories, content)
	}

	return memories, nil
}

// GetMemory gets a memory from the database.
func (db *DB) GetMemory(id int64) (string, error) {
	var content string
	err := db.QueryRow("SELECT content FROM memories WHERE id = ?", id).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}

// Entity represents an entity in the knowledge graph.
type Entity struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetEntities retrieves all entities from the database.
func (db *DB) GetEntities() ([]Entity, error) {
	rows, err := db.Query("SELECT id, name, type FROM entities")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var entity Entity
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.Type); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

// Relationship represents a relationship between two entities.
type Relationship struct {
	ID       int64  `json:"id"`
	SourceID int64  `json:"source_id"`
	TargetID int64  `json:"target_id"`
	Type     string `json:"type"`
}

// GetRelationships retrieves all relationships from the database.
func (db *DB) GetRelationships() ([]Relationship, error) {
	rows, err := db.Query("SELECT id, source_id, target_id, type FROM relationships")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relationships []Relationship
	for rows.Next() {
		var rel Relationship
		if err := rows.Scan(&rel.ID, &rel.SourceID, &rel.TargetID, &rel.Type); err != nil {
			return nil, err
		}
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

// AddRelationship adds a new relationship to the database.
func (db *DB) AddRelationship(sourceID, targetID int64, relType string) (int64, error) {
	result, err := db.Exec("INSERT INTO relationships (source_id, target_id, type) VALUES (?, ?, ?)", sourceID, targetID, relType)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetSnapshot gets a snapshot of the database.
func (db *DB) GetSnapshot() (*sql.DB, error) {
	return db.DB, nil
}

// RestoreSnapshot restores a snapshot of the database.
func (db *DB) RestoreSnapshot(snapshot *sql.DB) error {
	// This is a placeholder for a more complex snapshot restoration process.
	return nil
}

// GetDB returns the underlying database connection.
func (db *DB) GetDB() *sql.DB {
	return db.DB
}

// Close closes the database connection.
func (db *DB) Close() error {
	err := db.DB.Close()
	if err != nil {
		return err
	}
	if db.index != nil {
		if err := db.index.Close(); err != nil {
			return fmt.Errorf("failed to close bleve index: %w", err)
		}
	}
	return nil
}
