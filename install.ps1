# Requires PowerShell 5.1 or later

$ProjectName = "nodimus-memory"
$Repo = "wassmi/nodimus-memory" # Replace with your GitHub username/repo
$InstallDir = Join-Path $HOME ".nodimus-memory\bin"
$McpConfigDir = Join-Path $HOME ".nodimus-memory"
$McpConfigFile = Join-Path $McpConfigDir "mcp.json"

# --- Helper Functions ---
function Detect-OS {
    if ($IsWindows) { return "windows" }
    if ($IsLinux) { return "linux" }
    if ($IsOSX) { return "darwin" }
    return "unsupported"
}

function Detect-Arch {
    $arch = (Get-CimInstance Win32_Processor).Architecture
    # 0: x86, 9: x64 (AMD64), 5: ARM
    switch ($arch) {
        0 { return "amd64" } # Assuming x86 for 32-bit Windows, but we target amd64
        9 { return "amd64" }
        5 { return "arm64" } # Assuming ARM for Windows on ARM
        default {
            # For Linux/macOS, use [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
            $osArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
            if ($osArch -eq "X64") { return "amd64" }
            if ($osArch -eq "Arm64") { return "arm64" }
            return "unsupported"
        }
    }
}

function Get-LatestRelease {
    try {
        $uri = "https://api.github.com/repos/$Repo/releases/latest"
        $response = Invoke-RestMethod -Uri $uri -Headers @{"Accept"="application/vnd.github.v3+json"}
        return $response.tag_name
    } catch {
        Write-Error "Failed to get latest release tag: $($_.Exception.Message)"
        exit 1
    }
}

function Download-File {
    param (
        [string]$Url,
        [string]$OutputPath
    )
    Write-Host "Downloading $Url to $OutputPath..."
    try {
        Invoke-WebRequest -Uri $Url -OutFile $OutputPath -UseBasicParsing
    } catch {
        Write-Error "Failed to download file: $($_.Exception.Message)"
        exit 1
    }
}

# --- Main Installation Logic ---
$OS = Detect-OS
$ARCH = Detect-Arch

if ($OS -eq "unsupported" -or $ARCH -eq "unsupported") {
    Write-Error "Unsupported OS ($OS) or architecture ($ARCH)."
    exit 1
}

Write-Host "Detected OS: $OS, Architecture: $ARCH"

$LatestTag = Get-LatestRelease
if ([string]::IsNullOrEmpty($LatestTag)) {
    Write-Error "Could not determine latest release tag."
    exit 1
}
Write-Host "Latest release: $LatestTag"

$BinaryName = $ProjectName
$ArchiveExt = "tar.gz"
if ($OS -eq "windows") {
    $BinaryName = "${ProjectName}.exe"
    $ArchiveExt = "zip"
}

$DownloadUrl = "https://github.com/$Repo/releases/download/$LatestTag/${ProjectName}_${OS}_${ARCH}.${ArchiveExt}"
$TempArchive = Join-Path $env:TEMP "${ProjectName}_${LatestTag}.${ArchiveExt}"

Download-File -Url $DownloadUrl -OutputPath $TempArchive

Write-Host "Extracting $TempArchive..."
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

if ($ArchiveExt -eq "tar.gz") {
    # PowerShell doesn't have native tar.gz extraction. Use 7-Zip or similar if available, or rely on tar if present (e.g., Git Bash)
    # For simplicity, we'll assume tar is available on non-Windows or use Expand-Archive for zip
    if ($OS -ne "windows") {
        tar -xzf $TempArchive -C $InstallDir $BinaryName
    } else {
        Write-Error "Tar.gz extraction not natively supported on Windows PowerShell. Please extract manually or use a tool like 7-Zip."
        exit 1
    }
} else { # zip
    Expand-Archive -Path $TempArchive -DestinationPath $InstallDir -Force
}
Remove-Item $TempArchive

Write-Host "Installed $BinaryName to $(Join-Path $InstallDir $BinaryName)"

# Create or update mcp.json
Write-Host "Creating/updating $McpConfigFile..."
if (-not (Test-Path $McpConfigDir)) {
    New-Item -ItemType Directory -Path $McpConfigDir | Out-Null
}

$McpConfigContent = @{
    mcpServers = @{
        nodimus-memory = @{
            command = Join-Path $InstallDir $BinaryName
            args = @("mcp")
        }
    }
} | ConvertTo-Json -Depth 3

$McpConfigContent | Set-Content -Path $McpConfigFile
Write-Host "MCP configuration written to $McpConfigFile"

Write-Host "Installation complete!"
Write-Host "You can now add Nodimus Memory to your LLM CLI. For example:"
Write-Host "claude mcp add nodimus-memory ~/.nodimus-memory/bin/nodimus-memory mcp"
Write-Host "gemini mcp add nodimus-memory ~/.nodimus-memory/bin/nodimus-memory mcp"
Write-Host "cursor mcp add nodimus-memory ~/.nodimus-memory/bin/nodimus-memory mcp"
