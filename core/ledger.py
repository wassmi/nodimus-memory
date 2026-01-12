import asyncio
import json
import os
import math
import time
import uuid
from typing import Dict, Any, List, Optional
from cryptography.fernet import Fernet

# Graceful Failure: Try imports
# try:
#     import lancedb
#     import pyarrow as pa
#     LANCE_AVAILABLE = True
# except ImportError:
#     LANCE_AVAILABLE = False
#     print("Warning: lancedb not found. Falling back to JSONL.")
LANCE_AVAILABLE = False
print("Warning: lancedb disabled due to environment issues. Falling back to JSONL.")

try:
    import kuzu
    KUZU_AVAILABLE = True
except ImportError:
    KUZU_AVAILABLE = False
    print("Warning: kuzu not found. Falling back to JSONL.")

class Ledger:
    def __init__(self, db_path: str = "./data", encryption_key: Optional[bytes] = None):
        self.db_path = db_path
        os.makedirs(db_path, exist_ok=True)
        
        # Encryption
        if not encryption_key:
            self.key = Fernet.generate_key()
            print(f"Generated new encryption key: {self.key.decode()}")
        else:
            self.key = encryption_key
        self.cipher = Fernet(self.key)

        # Fallback
        self.fallback_file = os.path.join(db_path, "fallback.jsonl")

        # Initialize LanceDB
        self.lance_db = None
        self.episodes_table = None
        if LANCE_AVAILABLE:
            try:
                self.lance_db = lancedb.connect(os.path.join(db_path, "lancedb"))
                # Define schema if not exists (simplified for generation)
                # In real usage, we'd define a PyArrow schema
                self.episodes_table = self.lance_db.create_table(
                    "episodes",
                    schema=pa.schema([
                        pa.field("id", pa.string()),
                        pa.field("vector", pa.list_(pa.float32(), 1536)),
                        pa.field("content", pa.string()),
                        pa.field("metadata", pa.string()), # Encrypted JSON
                        pa.field("timestamp", pa.float64()),
                        pa.field("weight", pa.float64())
                    ]),
                    exist_ok=True
                )
            except Exception as e:
                print(f"LanceDB Init Failed: {e}")
                self.lance_db = None

        # Initialize Kuzu
        self.kuzu_db = None
        self.kuzu_conn = None
        if KUZU_AVAILABLE:
            try:
                self.kuzu_db = kuzu.Database(os.path.join(db_path, "kuzu"))
                self.kuzu_conn = kuzu.Connection(self.kuzu_db)
                self._init_kuzu_schema()
            except Exception as e:
                print(f"Kuzu Init Failed: {e}")
                self.kuzu_conn = None

    def _init_kuzu_schema(self):
        if not self.kuzu_conn:
            return
        # Create Node Tables
        try:
            self.kuzu_conn.execute("CREATE NODE TABLE Memory(id STRING, type STRING, weight DOUBLE, PRIMARY KEY (id))")
            self.kuzu_conn.execute("CREATE NODE TABLE File(path STRING, PRIMARY KEY (path))")
            self.kuzu_conn.execute("CREATE NODE TABLE Artifact(type STRING, content STRING, PRIMARY KEY (content))") # Content as PK is risky but simplifies
            # Create Rel Tables
            self.kuzu_conn.execute("CREATE REL TABLE RELATED(FROM Memory TO Memory)")
            self.kuzu_conn.execute("CREATE REL TABLE AFFECTS(FROM Memory TO File)")
            self.kuzu_conn.execute("CREATE REL TABLE CONTAINS(FROM File TO Artifact)")
        except Exception:
            pass # Tables likely exist

    def encrypt(self, data: str) -> str:
        return self.cipher.encrypt(data.encode()).decode()

    def decrypt(self, data: str) -> str:
        return self.cipher.decrypt(data.encode()).decode()

    async def add_episode(self, episode: Dict[str, Any]):
        # Encrypt content and metadata
        encrypted_content = self.encrypt(episode["content"])
        encrypted_metadata = self.encrypt(json.dumps(episode["metadata"]))
        
        record = {
            "id": episode["id"],
            "vector": episode["vector"],
            "content": encrypted_content,
            "metadata": encrypted_metadata,
            "timestamp": episode["timestamp"],
            "weight": episode["weight"]
        }

        # LanceDB Write
        if self.episodes_table:
            try:
                self.episodes_table.add([record])
            except Exception as e:
                print(f"LanceDB Write Error: {e}")
                await self._write_fallback(record)
        else:
            await self._write_fallback(record)

        # Kuzu Write (Graph)
        if self.kuzu_conn:
            try:
                # Create Memory Node
                self.kuzu_conn.execute(
                    "MERGE (m:Memory {id: $id, type: 'episode', weight: $weight})",
                    {"id": episode["id"], "weight": episode["weight"]}
                )
                # Link to File if exists in metadata
                root = episode["metadata"].get("root")
                if root:
                    self.kuzu_conn.execute(
                        "MERGE (f:File {path: $path})",
                        {"path": root}
                    )
                    self.kuzu_conn.execute(
                        "MATCH (m:Memory {id: $id}), (f:File {path: $path}) MERGE (m)-[:AFFECTS]->(f)",
                        {"id": episode["id"], "path": root}
                    )
            except Exception as e:
                print(f"Kuzu Write Error: {e}")

    async def _write_fallback(self, record: Dict[str, Any]):
        async with aiofiles.open(self.fallback_file, "a") as f:
            await f.write(json.dumps(record) + "\n")

    async def decay_loop(self):
        """
        Background loop to decay weights.
        node.weight *= exp(-0.01 * hours_since_last_access)
        """
        while True:
            await asyncio.sleep(3600) # Run every hour
            print("Running Decay Loop...")
            if self.kuzu_conn:
                try:
                    # Simplified decay: just multiply all weights
                    # In reality, we'd track access time.
                    decay_factor = math.exp(-0.01 * 1) 
                    self.kuzu_conn.execute(
                        f"MATCH (m:Memory) SET m.weight = m.weight * {decay_factor}"
                    )
                except Exception as e:
                    print(f"Decay Error: {e}")

import aiofiles
