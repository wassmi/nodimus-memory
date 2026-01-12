from typing import Any, List, Dict
from mcp.server.fastmcp import FastMCP
from core.ledger import Ledger

# Global ledger reference to be set by main
# In a more complex app, we'd use dependency injection or a class-based server
_ledger: Ledger = None

def set_ledger(ledger: Ledger):
    global _ledger
    _ledger = ledger

mcp = FastMCP("nodimus")

@mcp.tool()
async def get_memory_context(query: str, current_file: str, limit: int = 5) -> str:
    """
    Retrieves relevant memory context based on query and current file.
    Performs Vector Search + 1-hop Graph Traversal.
    """
    if not _ledger:
        return "Error: Ledger not initialized."

    results = []
    
    # 1. Vector Search (LanceDB)
    if _ledger.episodes_table:
        try:
            # Placeholder for embedding generation
            # In real app, use OpenAI or local model to embed `query`
            query_vector = [0.0] * 1536 
            
            search_res = _ledger.episodes_table.search(query_vector).limit(limit).to_list()
            for r in search_res:
                # Decrypt
                content = _ledger.decrypt(r["content"])
                results.append(f"- [Vector] {content} (Score: {r.get('_distance', 0)})")
                
                # 2. Graph Traversal (1-hop) (Kuzu)
                if _ledger.kuzu_conn:
                    try:
                        # Find related memories
                        q = f"""
                            MATCH (m:Memory {{id: '{r['id']}'}})-[:RELATED]->(related:Memory)
                            RETURN related.id
                        """
                        related = _ledger.kuzu_conn.execute(q)
                        while related.has_next():
                            row = related.get_next()
                            # Fetch related content (simplified, would need lookup)
                            results.append(f"  -> Related Memory: {row[0]}")
                    except Exception as e:
                        pass
        except Exception as e:
            results.append(f"Vector search failed: {e}")

    # Format Output
    markdown = "### Relevant Past Activity\n\n"
    if not results:
        markdown += "No relevant memories found."
    else:
        markdown += "\n".join(results)
        
    return markdown

def start_mcp_server():
    # This is usually handled by the MCP CLI or runner
    mcp.run()
