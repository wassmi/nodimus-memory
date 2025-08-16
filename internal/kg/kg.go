package kg

import (
	"encoding/json"
	"os"

	"github.com/wassmi/nodimus-memory/internal/storage"
)

// KGDB defines the database operations required by the kg package.
type KGDB interface {
	GetEntities() ([]storage.Entity, error)
	GetRelationships() ([]storage.Relationship, error)
}

// Generate generates a knowledge graph file in JSON-LD format.
func Generate(db KGDB, path string) error {
	entities, err := db.GetEntities()
	if err != nil {
		return err
	}

	relationships, err := db.GetRelationships()
	if err != nil {
		return err
	}

	graph := map[string]interface{}{
		"@context": "https://schema.org/",
		"@graph":   []interface{}{},
	}

	entityMap := make(map[int64]map[string]interface{})

	for _, entity := range entities {
		entityNode := map[string]interface{}{
			"@type": "Thing",
			"@id":   entity.ID,
			"name":  entity.Name,
			"type":  entity.Type,
		}
		graph["@graph"] = append(graph["@graph"].([]interface{}), entityNode)
		entityMap[entity.ID] = entityNode
	}

	for _, rel := range relationships {
		graph["@graph"] = append(graph["@graph"].([]interface{}), map[string]interface{}{
			"@type":       "Relationship",
			"@id":         rel.ID,
			"source":      rel.SourceID,
			"target":      rel.TargetID,
			"relationshipType": rel.Type,
		})
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(graph)
}