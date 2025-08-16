-- Nodimus-Memory MCP Database Schema

-- Stores individual memories or conversational turns
CREATE TABLE IF NOT EXISTS memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Stores unique, named entities (e.g., files, libraries, concepts)
CREATE TABLE IF NOT EXISTS entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL -- e.g., 'file', 'person', 'library'
);

-- Links memories to the entities mentioned within them, creating the graph
CREATE TABLE IF NOT EXISTS memory_entities (
    memory_id INTEGER NOT NULL,
    entity_id INTEGER NOT NULL,
    PRIMARY KEY (memory_id, entity_id),
    FOREIGN KEY (memory_id) REFERENCES memories (id) ON DELETE CASCADE,
    FOREIGN KEY (entity_id) REFERENCES entities (id) ON DELETE CASCADE
);

-- Stores relationships between entities
CREATE TABLE IF NOT EXISTS relationships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    FOREIGN KEY (source_id) REFERENCES entities (id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES entities (id) ON DELETE CASCADE
);