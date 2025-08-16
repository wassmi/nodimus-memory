# Nodimus Memory: The LLM's Second Brain

Nodimus Memory is designed for 100% local operation, ensuring your data remains private and on your machine.

Nodimus Memory is a specialized memory and knowledge management system designed to act as a "second brain" for Large Language Models (LLMs) and human users. It provides a robust, searchable, and extensible platform for storing, retrieving, and organizing information, enabling LLMs to access and leverage a persistent knowledge base beyond their immediate context.

## Features

*	**Persistent Memory Storage:** Stores "memories" (text content) and associated entities in a SQLite database.
*	**Advanced Full-Text Search:** Utilizes `bleve` for fast and powerful full-text search capabilities, allowing LLMs and users to retrieve relevant information efficiently.
*	**Asynchronous Knowledge Graph Generation:** Automatically generates a JSON-LD knowledge graph in the background, providing a structured representation of stored information for advanced reasoning and analysis.
*	**Structured Logging:** Implements robust logging using `lumberjack` for better monitoring, debugging, and auditing.
*	**Extensible Architecture:** Designed with modularity in mind, allowing for future enhancements such as vector embeddings, more sophisticated relationship management, and diverse data sources.

## Why Nodimus Memory?

LLMs, while powerful, often lack persistent memory and struggle with long-term knowledge retention. Nodimus Memory addresses this by providing:

*	**Enhanced Context:** LLMs can query Nodimus Memory to retrieve relevant past interactions, facts, or learned information, significantly expanding their contextual understanding.
*	**Reduced Hallucinations:** By grounding LLM responses in a verifiable knowledge base, Nodimus Memory helps reduce factual inaccuracies and "hallucinations."
*	**Personalized Experiences:** Over time, Nodimus Memory can accumulate user-specific knowledge, enabling LLMs to provide more personalized and accurate responses.
*	**Human-in-the-Loop Knowledge Curation:** Users can directly add and manage memories, curating the knowledge base that informs the LLM.

## Getting Started

### Quick Start

Get Nodimus Memory up and running with a single command. The installer will detect your OS and architecture, download the appropriate binary, and set up the necessary configuration.

**Linux/macOS:**

```bash
curl -sSf https://nodimus.com/install.sh | sh
```

**Windows (PowerShell):**

```powershell
iwr https://nodimus.com/install.ps1 -useb | iex
```

After installation, you can add Nodimus Memory as an MCP server to your favorite LLM CLI. For example:

```bash
# For Claude CLI
claude mcp add nodimus "$(cat ~/.nodimus/mcp.json)"
where ~/.nodimus/mcp.json looks like:

```json
{
  "command": "~/.nodimus/bin/nodimus",
  "args": ["mcp"]
}
```

Quick sanity check:

```bash
# Confirm the MCP descriptor is valid JSON
jq . ~/.nodimus/mcp.json
```
If it prints without error, the LLM CLI will be able to launch nodimus mcp over STDIO.
(You can also hand-edit ~/.claude/mcp.json if you prefer.)

# For Gemini CLI
gemini mcp add nodimus "$(cat ~/.nodimus/mcp.json)"
~/.nodimus/mcp.json is the same JSON blob as above.

#### Cursor
Cursor reads the MCP list from its settings file, not via a CLI sub-command.
Add the following stanza to ~/.cursor/settings.json (create the file if it doesn’t exist):

```jsonc
{
  "mcp_servers": {
    "nodimus_memory": {
      "command": "~/.nodimus/bin/nodimus",
      "args": ["mcp"]
    }
  }
}
```

Quick sanity check:

```bash
# Confirm the MCP descriptor is valid JSON
jq . ~/.nodimus/mcp.json
```
If it prints without error, the LLM CLI will be able to launch nodimus mcp over STDIO.
```

### Configuration

Nodimus Memory uses a `config.toml` file for configuration. A default `config.toml` is provided:

```toml
[server]
port = 8080
bind = "127.0.0.1"
timeout = 30

[storage]
data_dir = "~/.nodimus"

[logger]
level = "info"
file = "audit/nodimus.log"
max_size = 50
max_backups = 3
max_age = 30
compress = true
```

You can modify this file to change the server port, bind address, data directory, and logging settings.

## Usage: Interacting with Nodimus Memory

Nodimus Memory exposes a JSON-RPC 2.0 API for interaction, making it easy for LLMs and other applications to integrate.

### API Endpoints

*	`/rpc`: JSON-RPC 2.0 endpoint

### Methods

#### `memory.AddMemory`

Adds a new memory to the system.

**Example (using `curl`):**

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc": "2.0", "method": "memory.AddMemory", "params": [{"content": "The capital of France is Paris.", "entities": ["France", "Paris"]}], "id": 1}' \
     http://127.0.0.1:8080/rpc
```

**Example (LLM Integration - conceptual):**

```python
# Python example for an LLM agent to add memory
import requests

def add_memory_to_nodimus(content, entities):
    url = "http://127.0.0.1:8080/rpc"
    headers = {"Content-Type": "application/json"}
    payload = {
        "jsonrpc": "2.0",
        "method": "memory.AddMemory",
        "params": [{
            "content": content,
            "entities": entities
        }],
        "id": 1
    }
    response = requests.post(url, json=payload, headers=headers)
    response.raise_for_status()
    return response.json()

# LLM decides to store a new fact
llm_generated_content = "The Eiffel Tower is in Paris."
llm_extracted_entities = ["Eiffel Tower", "Paris"]

response = add_memory_to_nodimus(llm_generated_content, llm_extracted_entities)
print(f"Memory added with ID: {response['result']['id']}")
```

#### `memory.SearchMemory`

Searches for memories based on a query.

**Example (using `curl`):**

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc": "2.0", "method": "memory.SearchMemory", "params": [{"query": "capital of France"}], "id": 1}' \
     http://127.0.0.1:8080/rpc
```

**Example (LLM Integration - conceptual):**

```python
# Python example for an LLM agent to search memory
import requests

def search_memory_in_nodimus(query):
    url = "http://127.0.0.1:8080/rpc"
    headers = {"Content-Type": "application/json"}
    payload = {
        "jsonrpc": "2.0",
        "method": "memory.SearchMemory",
        "params": [{
            "query": query
        }],
        "id": 1
    }
    response = requests.post(url, json=payload, headers=headers)
    response.raise_for_status()
    return response.json()

# LLM needs to retrieve information
llm_query = "What is the capital of France?"
search_results = search_memory_in_nodimus(llm_query)

if search_results['result']['results']:
    print("Found memories:")
    for memory in search_results['result']['results']:
        print(f"- {memory}")
else:
    print("No relevant memories found.")
```

#### `memory.GetContext`

Retrieves the content of a specific memory by its ID.

**Example (using `curl`):

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc": "2.0", "method": "memory.GetContext", "params": [{"id": 123}], "id": 1}' \
     http://127.0.0.1:8080/rpc
```

### LLM Integration Examples

#### Gemini CLI

To integrate Nodimus Memory with Gemini CLI, you'll need to configure a custom tool. Add the following to your Gemini CLI settings file (e.g., `~/.gemini-cli/settings.json`):

```json
{
  "mcp_servers": {
    "nodimus_memory": {
      "type": "http",
      "url": "http://127.0.0.1:8080/rpc",
      "api_key": ""
    }
  },
  "tools": [
    {
      "name": "nodimus_add_memory",
      "description": "Adds a new memory to Nodimus Memory.",
      "mcp_server": "nodimus_memory",
      "method": "memory.AddMemory",
      "parameters": {
        "content": {
          "type": "string",
          "description": "The content of the memory."
        },
        "entities": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "description": "A list of entities associated with the memory."
        }
      }
    },
    {
      "name": "nodimus_search_memory",
      "description": "Searches for memories in Nodimus Memory.",
      "mcp_server": "nodimus_memory",
      "method": "memory.SearchMemory",
      "parameters": {
        "query": {
          "type": "string",
          "description": "The search query."
        }
      }
    },
    {
      "name": "nodimus_get_context",
      "description": "Retrieves the content of a specific memory by its ID from Nodimus Memory.",
      "mcp_server": "nodimus_memory",
      "method": "memory.GetContext",
      "parameters": {
        "id": {
          "type": "integer",
          "description": "The ID of the memory to retrieve."
        }
      }
    }
  ]
}
```

Once configured, you can use these tools within Gemini CLI:

```
/tool nodimus_add_memory content="The quick brown fox jumps over the lazy dog." entities=["fox", "dog"]
/tool nodimus_search_memory query="brown fox"
```

#### Claude (Conceptual Integration)

For Claude, you would typically integrate by making HTTP requests to the Nodimus Memory server from your application or script that interacts with Claude. This often involves a custom function or service that acts as an intermediary.

Here's a conceptual Python example of how you might integrate with Claude, assuming you have a way to send messages to Claude and receive responses (e.g., via an SDK or API):

```python
import requests
import json

NODIMUS_RPC_URL = "http://127.0.0.1:8080/rpc"

def call_nodimus_rpc(method, params):
    headers = {"Content-Type": "application/json"}
    payload = {
        "jsonrpc": "2.0",
        "method": method,
        "params": [params],
        "id": 1
    }
    try:
        response = requests.post(NODIMUS_RPC_URL, json=payload, headers=headers)
        response.raise_for_status()
        return response.json().get("result")
    except requests.exceptions.RequestException as e:
        print(f"Error calling Nodimus Memory RPC: {e}")
        return None

def claude_add_memory(content, entities):
    """Function to be called by your Claude integration to add memory."""
    print(f"Claude is adding memory: {content} with entities {entities}")
    result = call_nodimus_rpc("memory.AddMemory", {"content": content, "entities": entities})
    if result:
        print(f"Memory added with ID: {result.get('id')}")
    return result

def claude_search_memory(query):
    """Function to be called by your Claude integration to search memory."""
    print(f"Claude is searching memory for query: {query}")
    result = call_nodimus_rpc("memory.SearchMemory", {"query": query})
    if result and result.get("results"):
        print("Found memories:")
        for memory in result["results"]:
            print(f"- {memory}")
        return result["results"]
    else:
        print("No relevant memories found.")
        return []

# Example of how Claude might use these functions (within your application logic)
# This is pseudo-code, as actual Claude integration depends on your specific setup.

# Scenario 1: Claude learns a new fact and stores it
# if claude_response_indicates_new_fact:
#     fact_content = "The capital of Canada is Ottawa."
#     fact_entities = ["Canada", "Ottawa"]
#     claude_add_memory(fact_content, fact_entities)

# Scenario 2: Claude needs to retrieve information
# if claude_needs_to_answer_question:
#     question_query = "What is the capital of Canada?"
#     retrieved_memories = claude_search_memory(question_query)
#     # Claude would then use retrieved_memories to formulate its answer
#     if retrieved_memories:
#         print(f"Claude uses retrieved memories to answer: {retrieved_memories[0]}")
```

## Deployment


For production deployments, consider containerization using Docker. A `Dockerfile` will be provided in future releases to simplify deployment to various environments.

## Development

If you wish to contribute or build from source:

### Prerequisites

*	Go (version 1.20 or higher)

### Installation

1.	**Clone the repository:**
	```bash
	git clone https://github.com/nodimus/nodimus-memory.git
	cd nodimus-memory
	```

2.	**Install dependencies:**
	```bash
	go mod tidy
	```

### Running Tests

```bash
go test ./...
```

### Building the Executable

```bash
go build -o nodimus-memory ./cmd/nodimus
```

## Future Enhancements

*	**Vector Embeddings:** Integrate vector embeddings for semantic search and similarity-based retrieval, allowing LLMs to find conceptually related memories even without exact keyword matches.
*	**More Sophisticated Relationship Management:** Enhance the knowledge graph with richer relationship types and inference capabilities.
*	**Diverse Data Sources:** Support ingesting memories from various sources (e.g., web pages, documents, chat logs).
*	**Authentication and Authorization:** Implement security measures for multi-user environments.
*	**Scalability:** Explore options for distributed deployment and horizontal scaling.

## Contributing

Contributions are welcome! Please see the `CONTRIBUTING.md` (to be created) for guidelines.

## License

This project is licensed under the MIT License.