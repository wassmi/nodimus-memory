package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/wassmi/nodimus-memory/internal/config"
	"github.com/wassmi/nodimus-memory/internal/kg"
	"github.com/wassmi/nodimus-memory/internal/logger"
	"github.com/wassmi/nodimus-memory/internal/seccomp"
	"github.com/wassmi/nodimus-memory/internal/server"
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

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.toml", "path to config file")
	rootCmd.AddCommand(mcpCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type CommonLogger interface {
	Fatalf(format string, v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type ConfigProvider interface {
	ExpandDataDir() (string, error)
}

type DBProvider interface {
	NewDB(dataSourceName string) (*storage.DB, error)
}

func setupCommon(log CommonLogger, cfg ConfigProvider, dbProvider DBProvider) (*storage.DB, string, error) {
	if err := seccomp.LoadProfile(); err != nil {
		return nil, "", fmt.Errorf("failed to load seccomp profile: %w", err)
	}
	dataDir, err := cfg.ExpandDataDir()
	if err != nil {
		return nil, "", fmt.Errorf("failed to expand data dir: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create data dir: %w", err)
	}
	db, err := dbProvider.NewDB(filepath.Join(dataDir, "nodimus-memory.db"))
	if err != nil {
		return nil, "", fmt.Errorf("failed to open database: %w", err)
	}
	if db == nil {
		return nil, "", errors.New("database connection is nil")
	}
	if err := db.Migrate(); err != nil {
		return nil, "", fmt.Errorf("failed to migrate database: %w", err)
	}
	if err := snapshot.IntegrityCheck(db); err != nil {
		return nil, "", fmt.Errorf("database integrity check failed: %w", err)
	}
	if err := kg.Generate(db, filepath.Join(dataDir, "knowledge-graph.jsonld")); err != nil {
		log.Printf("failed to generate knowledge graph: %v\n", err)
	}
	return db, dataDir, nil
}

type realDBProvider struct{}

func (r *realDBProvider) NewDB(dataSourceName string) (*storage.DB, error) {
	return storage.NewDB(dataSourceName)
}

func runHTTPServer() {
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err)
	}
	appLogger := logger.New(cfg.Logger)
	db, dataDir, err := setupCommon(appLogger, cfg, &realDBProvider{})
	if err != nil {
		appLogger.Fatalf("Setup failed: %v", err)
	}
	defer db.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	mcpService := &server.MemoryService{DB: db, DataDir: dataDir, Log: appLogger}
	mcpServer := server.NewServer(cfg.Server.Port, cfg.Server.Bind, cfg.Server.Timeout, mcpService)
	go func() {
		appLogger.Printf("MCP server listening on %s:%d\n", cfg.Server.Bind, cfg.Server.Port)
		if err := mcpServer.Start(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("MCP server failed: %v\n", err)
		}
	}()

	metricsServer := server.NewMetricsServer(9090, "127.0.0.1")
	go func() {
		appLogger.Printf("Metrics server listening on http://127.0.0.1:9090/metrics\n")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Printf("Metrics server failed: %v\n", err)
		}
	}()

	snapshotter := snapshot.NewSnapshotter()
	if err := snapshotter.Start(db, dataDir); err != nil {
		appLogger.Fatalf("failed to start snapshotter: %v\n", err)
	}
	defer snapshotter.Stop()

	<-	sigChan
	appLogger.Println("\nShutting down servers...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := mcpServer.Stop(ctx); err != nil {
		appLogger.Printf("failed to stop MCP server: %v\n", err)
	}
	appLogger.Println("Servers stopped.")
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      *json.RawMessage `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      *json.RawMessage `json:"id"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func runStdioServer() {
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err)
	}
	appLogger := logger.New(cfg.Logger)
	db, dataDir, err := setupCommon(appLogger, cfg, &realDBProvider{})
	if err != nil {
		appLogger.Fatalf("Setup failed: %v", err)
	}
	defer db.Close()

	mcpService := &server.MemoryService{DB: db, DataDir: dataDir, Log: appLogger}
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	appLogger.Println("MCP stdio server started.")
	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					appLogger.Printf("Error reading from stdin: %v\n", err)
				}
				return
			}
			var req JSONRPCRequest
			if err := json.Unmarshal(line, &req); err != nil {
				// Handle parse error
				continue
			}
			var resp JSONRPCResponse
			resp.JSONRPC = "2.0"
			resp.ID = req.ID
			switch req.Method {
			case "memory.AddMemory":
				var params server.AddMemoryRequest
				if err := json.Unmarshal(req.Params, &params); err == nil {
					var reply server.AddMemoryResponse
					if err := mcpService.AddMemory(nil, &params, &reply); err == nil {
						resp.Result = reply
					}
				}
			// ... other cases
			}
			writeResponse(writer, resp, appLogger)
		}
	}()
	<-	sigChan
	appLogger.Println("Exiting stdio server.")
}

func writeResponse(writer *bufio.Writer, resp JSONRPCResponse, log CommonLogger) {
	respBytes, _ := json.Marshal(resp)
	writer.Write(respBytes)
	writer.WriteString("\n")
	writer.Flush()
}