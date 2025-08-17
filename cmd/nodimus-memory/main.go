package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/wassmi/nodimus-memory/internal/config"
	"github.com/wassmi/nodimus-memory/internal/kg"
	"github.com/wassmi/nodimus-memory/internal/logger"
	"github.com/wassmi/nodimus-memory/internal/snapshot"
	"github.com/wassmi/nodimus-memory/internal/storage"
	"github.com/spf13/cobra"
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:   "nodimus-memory",
		Short: "Nodimus Memory: The LLM's Second Brain",
		Long:  `Nodimus Memory is a specialized memory and knowledge management system designed to act as a "second brain" for Large Language Models (LLMs) and human users.`,
		Run: func(cmd *cobra.Command, args []string) {
			runHTTPServer()
		},
	}
	mcpCmd = &cobra.Command{
		Use:    "mcp",
		Short:  "Starts the MCP server over stdio (hidden)",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			runStdioServer()
		},
	}
)

// createFailsafeLogger creates a logger that writes to a known location before
// the main configuration is loaded. This is crucial for debugging startup errors.
func createFailsafeLogger() (*log.Logger, *os.File) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Could not get home directory for failsafe logger: %v", err)
		return log.New(io.Discard, "", 0), nil
	}
	logDir := filepath.Join(homeDir, ".nodimus-memory")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Could not create directory for failsafe logger: %v", err)
		return log.New(io.Discard, "", 0), nil
	}
	logPath := filepath.Join(logDir, "startup.log")
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Could not open failsafe log file: %v", err)
		return log.New(io.Discard, "", 0), nil
	}
	return log.New(f, "STARTUP: ", log.LstdFlags|log.Lshortfile), f
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "path to config file (default is ~/.nodimus-memory/config.toml)")
	rootCmd.AddCommand(mcpCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func ensureConfig(userConfigPath string, failsafeLog *log.Logger) (*config.Config, error) {
	failsafeLog.Println("Ensuring configuration exists...")
	var finalConfigPath string
	if userConfigPath != "" {
		finalConfigPath = userConfigPath
		failsafeLog.Printf("Using user-provided config path: %s", finalConfigPath)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not get user home directory: %w", err)
		}
		configDir := filepath.Join(homeDir, ".nodimus-memory")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("could not create config directory %s: %w", configDir, err)
		}
		finalConfigPath = filepath.Join(configDir, "config.toml")
		failsafeLog.Printf("Using default config path: %s", finalConfigPath)
	}

	if _, err := os.Stat(finalConfigPath); os.IsNotExist(err) {
		failsafeLog.Printf("Config file not found. Creating default at %s", finalConfigPath)
		defaultConfig := config.Default()
		if err := defaultConfig.Save(finalConfigPath); err != nil {
			return nil, fmt.Errorf("could not save default config file: %w", err)
		}
	}
	failsafeLog.Println("Loading configuration...")
	return config.Load(finalConfigPath)
}

type CommonLogger interface {
	Fatalf(format string, v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type DBProvider interface {
	NewDB(dataSourceName string) (*storage.DB, error)
}

func setupCommon(log CommonLogger, cfg *config.Config, dbProvider DBProvider, failsafeLog *log.Logger) (*storage.DB, string, error) {
	failsafeLog.Println("Setting up common components...")
	dataDir, err := cfg.ExpandDataDir()
	if err != nil {
		return nil, "", fmt.Errorf("failed to expand data dir: %w", err)
	}
	failsafeLog.Printf("Data directory expanded to: %s", dataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create data dir: %w", err)
	}
	dbPath := filepath.Join(dataDir, "nodimus-memory.db")
	failsafeLog.Printf("Opening database at: %s", dbPath)
	db, err := dbProvider.NewDB(dbPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open database: %w", err)
	}
	if db == nil {
		return nil, "", errors.New("database connection is nil")
	}
	failsafeLog.Println("Migrating database...")
	if err := db.Migrate(); err != nil {
		return nil, "", fmt.Errorf("failed to migrate database: %w", err)
	}
	failsafeLog.Println("Running integrity check...")
	if err := snapshot.IntegrityCheck(db); err != nil {
		return nil, "", fmt.Errorf("database integrity check failed: %w", err)
	}
	kgPath := filepath.Join(dataDir, "knowledge-graph.jsonld")
	failsafeLog.Printf("Generating knowledge graph at: %s", kgPath)
	if err := kg.Generate(db, kgPath); err != nil {
		log.Printf("failed to generate knowledge graph: %v\n", err)
	}
	failsafeLog.Println("Common setup complete.")
	return db, dataDir, nil
}

type realDBProvider struct{}

func (r *realDBProvider) NewDB(dataSourceName string) (*storage.DB, error) {
	return storage.NewDB(dataSourceName)
}

func runHTTPServer() {
	failsafeLog, f := createFailsafeLogger()
	if f != nil {
		defer f.Close()
	}
	cfg, err := ensureConfig(configFile, failsafeLog)
	if err != nil {
		log.Fatalf("failed to load or create config: %v\n", err)
	}
	dataDir, err := cfg.ExpandDataDir()
	if err != nil {
		log.Fatalf("failed to expand data dir: %v\n", err)
	}
	appLogger := logger.New(cfg.Logger, dataDir)
	db, _, err := setupCommon(appLogger, cfg, &realDBProvider{}, failsafeLog)
	if err != nil {
		appLogger.Fatalf("Setup failed: %v", err)
	}
	defer db.Close()
	// ... (rest of the function is the same)
}

// ... (JSON-RPC types remain the same)

func runStdioServer() {
	failsafeLog, f := createFailsafeLogger()
	if f != nil {
		defer f.Close()
	}
	failsafeLog.Println("--- Stdio server starting ---")

	cfg, err := ensureConfig(configFile, failsafeLog)
	if err != nil {
		failsafeLog.Printf("ERROR: failed to ensure config: %v", err)
		fmt.Fprintf(os.Stderr, "failed to load or create config: %v\n", err)
		os.Exit(1)
	}
	dataDir, err := cfg.ExpandDataDir()
	if err != nil {
		failsafeLog.Printf("ERROR: failed to expand data dir: %v", err)
		fmt.Fprintf(os.Stderr, "failed to expand data dir: %v\n", err)
		os.Exit(1)
	}
	appLogger := logger.New(cfg.Logger, dataDir)
	db, _, err := setupCommon(appLogger, cfg, &realDBProvider{}, failsafeLog)
	if err != nil {
		failsafeLog.Printf("ERROR: setupCommon failed: %v", err)
		fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	reader := bufio.NewReader(os.Stdin)

	failsafeLog.Println("Entering main JSON-RPC read loop...")
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				failsafeLog.Println("EOF received on stdin. Exiting cleanly.")
			} else {
				failsafeLog.Printf("ERROR: Error reading from stdin: %v", err)
			}
			return // Exit on any read error, including EOF
		}
		failsafeLog.Printf("Received line from stdin: %s", string(line))
		// ... (rest of the loop is the same)
	}
}

// ... (writeResponse remains the same)
