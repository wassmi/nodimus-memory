#!/bin/sh

set -eu

PROJECT_NAME="nodimus-memory"
REPO="nodimus/nodimus-memory" # Replace with your GitHub username/repo
INSTALL_DIR="$HOME/.nodimus/bin"
MCP_CONFIG_DIR="$HOME/.nodimus"
MCP_CONFIG_FILE="$MCP_CONFIG_DIR/mcp.json"

# --- Helper Functions ---
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        CYGWIN*|MINGW32*|MSYS*|MINGW*) echo "windows";;
        *)          echo "unsupported"
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64)     echo "amd64";;
        arm64)      echo "arm64";;
        aarch64)    echo "arm64";;
        *)          echo "unsupported"
    esac
}

get_latest_release() {
    curl --silent "https://api.github.com/repos/$REPO/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

download_file() {
    local url="$1"
    local output="$2"
    echo "Downloading $url to $output..."
    curl --fail --location --progress-bar "$url" --output "$output"
}

# --- Main Installation Logic ---
OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
    echo "Error: Unsupported OS ($OS) or architecture ($ARCH)."
    exit 1
fi

echo "Detected OS: $OS, Architecture: $ARCH"

LATEST_TAG=$(get_latest_release)
if [ -z "$LATEST_TAG" ]; then
    echo "Error: Could not determine latest release tag."
    exit 1
fi
echo "Latest release: $LATEST_TAG"

BINARY_NAME="$PROJECT_NAME"
ARCHIVE_EXT="tar.gz"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${PROJECT_NAME}.exe"
    ARCHIVE_EXT="zip"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/${PROJECT_NAME}_${OS}_${ARCH}.${ARCHIVE_EXT}"
TEMP_ARCHIVE="/tmp/${PROJECT_NAME}_${LATEST_TAG}.${ARCHIVE_EXT}"

download_file "$DOWNLOAD_URL" "$TEMP_ARCHIVE"

echo "Extracting $TEMP_ARCHIVE..."
mkdir -p "$INSTALL_DIR"
if [ "$ARCHIVE_EXT" = "tar.gz" ]; then
    tar -xzf "$TEMP_ARCHIVE" -C "$INSTALL_DIR" "$BINARY_NAME"
else # zip
    unzip -o "$TEMP_ARCHIVE" -d "$INSTALL_DIR" "$BINARY_NAME"
fi
rm "$TEMP_ARCHIVE"

# Ensure the binary is executable
chmod +x "$INSTALL_DIR/$BINARY_NAME"
echo "Installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME"

# Create or update mcp.json
echo "Creating/updating $MCP_CONFIG_FILE..."
mkdir -p "$MCP_CONFIG_DIR"
cat << EOF > "$MCP_CONFIG_FILE"
{
  "mcpServers": {
    "nodimus": {
      "command": "$INSTALL_DIR/$BINARY_NAME",
      "args": ["mcp"]
    }
  }
}
EOF
echo "MCP configuration written to $MCP_CONFIG_FILE"

echo "Installation complete!"
echo "You can now add Nodimus Memory to your LLM CLI. For example:"
echo "claude mcp add nodimus ~/.nodimus/bin/nodimus mcp"
echo "gemini mcp add nodimus ~/.nodimus/bin/nodimus mcp"
echo "cursor mcp add nodimus ~/.nodimus/bin/nodimus mcp"