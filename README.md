# Nodimus Memory: The LLM's Second Brain

Nodimus Memory is designed for 100% local operation, ensuring your data remains private and on your machine.

Nodimus Memory is a specialized memory and knowledge management system designed to act as a "second brain" for Large Language Models (LLMs) and human users. It provides a robust, searchable, and extensible platform for storing, retrieving, and organizing information, enabling LLMs to access and leverage a persistent knowledge base beyond their immediate context.

## Getting Started

### Step 1: Install Nodimus Memory

Run the following command in your terminal. The installer will detect your OS and architecture, download the appropriate binary, and install it.

**Linux/macOS:**
```bash
curl -sSf https://raw.githubusercontent.com/wassmi/nodimus-memory/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
iwr https://raw.githubusercontent.com/wassmi/nodimus-memory/main/install.ps1 -useb | iex
```

The first time you run `nodimus-memory`, it will automatically create a default configuration file at `~/.nodimus-memory/config.toml`.

### Step 2: Connect to your LLM CLI

After installation, you can connect Nodimus Memory to your favorite LLM CLI.

**For Gemini CLI:**
```bash
gemini mcp add nodimus-memory nodimus-memory mcp
```

**For other CLIs (e.g., Claude, Cursor):**
You can find the command to run the server in the `command` field of the `~/.nodimus-memory/mcp.json` file that is automatically created.

## Development

If you wish to contribute or build from source:

### Prerequisites
*	Go (version 1.24.5 or higher)

### Installation
1.	**Clone the repository:**
	```bash
	git clone https://github.com/wassmi/nodimus-memory.git
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
go build -o nodimus-memory ./cmd/nodimus-memory
```
