package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/i2y/d2mcp/internal/infrastructure/d2"
	"github.com/i2y/d2mcp/internal/infrastructure/mcp"
	"github.com/i2y/d2mcp/internal/presentation/handler"
	"github.com/i2y/d2mcp/internal/usecase"
)

const (
	// ServerName is the name of the MCP server.
	ServerName = "d2mcp"
	// ServerVersion is the version of the MCP server.
	ServerVersion = "0.5.0"
)

func main() {
	// Disable D2 logs.
	os.Setenv("D2_LOG_LEVEL", "NONE")

	// Parse command line flags.
	var (
		transport         string
		addr              string
		baseURL           string
		basePath          string
		keepAlive         int
		endpointPath      string
		heartbeatInterval int
		stateless         bool
	)
	flag.StringVar(&transport, "transport", "sse", "Transport mode: stdio, sse, or streamable")
	flag.StringVar(&addr, "addr", ":3000", "Address to listen on for SSE/Streamable HTTP transport (e.g., :3000)")
	flag.StringVar(&baseURL, "base-url", "", "Base URL for SSE transport (e.g., http://localhost:3000)")
	flag.StringVar(&basePath, "base-path", "/mcp", "Base path for SSE endpoints")
	flag.IntVar(&keepAlive, "keep-alive", 30, "Keep-alive interval in seconds for SSE")
	flag.StringVar(&endpointPath, "endpoint-path", "/mcp", "Endpoint path for Streamable HTTP transport")
	flag.IntVar(&heartbeatInterval, "heartbeat-interval", 30, "Heartbeat interval in seconds for Streamable HTTP")
	flag.BoolVar(&stateless, "stateless", false, "Enable stateless mode for Streamable HTTP")
	flag.Parse()

	// Validate transport mode.
	if transport != "stdio" && transport != "sse" && transport != "streamable" {
		fmt.Fprintf(os.Stderr, "Invalid transport mode: %s. Must be 'stdio', 'sse', or 'streamable'\n", transport)
		os.Exit(1)
	}

	// Set up logging based on transport mode
	if transport == "stdio" {
		// In STDIO mode, log to file to avoid any interference with stdio communication
		logFile, err := os.OpenFile("/tmp/d2mcp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// If we can't open log file, just disable logging
			log.SetOutput(io.Discard)
		} else {
			log.SetOutput(logFile)
			log.Printf("D2MCP server starting in stdio mode, logging to %s", logFile.Name())
		}
	} else {
		// For other transports, stderr is fine
		log.SetOutput(os.Stderr)
	}

	// Log transport mode
	if transport == "sse" {
		log.Printf("Starting in SSE mode on %s", addr)
	} else if transport == "streamable" {
		log.Printf("Starting in Streamable HTTP mode on %s", addr)
	}

	// Create context.
	ctx := context.Background()

	// Initialize repository.
	oracleRepo := d2.NewD2OracleRepository()

	// Initialize usecases.
	diagramUseCase := usecase.NewDiagramUseCase(oracleRepo)
	oracleUseCase := usecase.NewOracleUseCase(oracleRepo)

	// Initialize MCP server.
	server, err := mcp.NewServer(ServerName, ServerVersion)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Configure transport
	switch transport {
	case "stdio":
		server.WithTransport(mcp.TransportStdio)
	case "sse":
		// Auto-generate base URL if not provided
		if baseURL == "" {
			if addr[0] == ':' {
				baseURL = fmt.Sprintf("http://localhost%s", addr)
			} else {
				baseURL = fmt.Sprintf("http://%s", addr)
			}
		}
		
		sseConfig := &mcp.SSEConfig{
			Addr:              addr,
			BaseURL:           baseURL,
			StaticBasePath:    basePath,
			KeepAliveInterval: time.Duration(keepAlive) * time.Second,
		}
		server.WithTransport(mcp.TransportSSE).WithSSEConfig(sseConfig)
		log.Printf("SSE endpoints will be available at:")
		log.Printf("  SSE: %s%s/sse", baseURL, basePath)
		log.Printf("  Messages: %s%s/message", baseURL, basePath)
	case "streamable":
		streamableConfig := &mcp.StreamableHTTPConfig{
			Addr:              addr,
			EndpointPath:      endpointPath,
			HeartbeatInterval: time.Duration(heartbeatInterval) * time.Second,
			Stateless:         stateless,
		}
		server.WithTransport(mcp.TransportStreamableHTTP).WithStreamableHTTPConfig(streamableConfig)
		log.Printf("Streamable HTTP endpoint will be available at:")
		log.Printf("  Endpoint: http://localhost%s%s", addr, endpointPath)
		if stateless {
			log.Printf("  Mode: Stateless")
		} else {
			log.Printf("  Mode: Stateful")
		}
	}

	// Initialize handlers.
	createHandler := handler.NewCreateHandler(diagramUseCase)
	exportHandler := handler.NewExportHandler(diagramUseCase)
	saveHandler := handler.NewSaveHandler(diagramUseCase)

	// Initialize Oracle handlers.
	oracleCreateHandler := handler.NewOracleCreateHandler(oracleUseCase)
	oracleSetHandler := handler.NewOracleSetHandler(oracleUseCase)
	oracleDeleteHandler := handler.NewOracleDeleteHandler(oracleUseCase)
	oracleMoveHandler := handler.NewOracleMoveHandler(oracleUseCase)
	oracleRenameHandler := handler.NewOracleRenameHandler(oracleUseCase)
	oracleGetHandler := handler.NewOracleGetHandler(oracleUseCase)
	oracleSerializeHandler := handler.NewOracleSerializeHandler(oracleUseCase)

	// Register tools.
	if err := server.RegisterTool(createHandler.GetTool(), createHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register create tool: %v", err)
	}
	if err := server.RegisterTool(exportHandler.GetTool(), exportHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register export tool: %v", err)
	}
	if err := server.RegisterTool(saveHandler.GetTool(), saveHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register save tool: %v", err)
	}

	// Register Oracle tools.
	if err := server.RegisterTool(oracleCreateHandler.GetTool(), oracleCreateHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register oracle create tool: %v", err)
	}
	if err := server.RegisterTool(oracleSetHandler.GetTool(), oracleSetHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register oracle set tool: %v", err)
	}
	if err := server.RegisterTool(oracleDeleteHandler.GetTool(), oracleDeleteHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register oracle delete tool: %v", err)
	}
	if err := server.RegisterTool(oracleMoveHandler.GetTool(), oracleMoveHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register oracle move tool: %v", err)
	}
	if err := server.RegisterTool(oracleRenameHandler.GetTool(), oracleRenameHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register oracle rename tool: %v", err)
	}
	if err := server.RegisterTool(oracleGetHandler.GetTool(), oracleGetHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register oracle get tool: %v", err)
	}
	if err := server.RegisterTool(oracleSerializeHandler.GetTool(), oracleSerializeHandler.GetHandler()); err != nil {
		log.Fatalf("Failed to register oracle serialize tool: %v", err)
	}

	// Start the server.
	log.Printf("Starting %s v%s MCP server...", ServerName, ServerVersion)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
