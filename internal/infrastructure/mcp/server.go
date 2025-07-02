package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TransportType represents the type of transport to use.
type TransportType string

const (
	// TransportStdio uses standard input/output.
	TransportStdio TransportType = "stdio"
	// TransportSSE uses Server-Sent Events over HTTP.
	TransportSSE TransportType = "sse"
	// TransportStreamableHTTP uses Streamable HTTP protocol.
	TransportStreamableHTTP TransportType = "streamable"
)

// SSEConfig contains configuration for SSE transport.
type SSEConfig struct {
	Addr               string
	BaseURL            string
	StaticBasePath     string
	KeepAliveInterval  time.Duration
}

// StreamableHTTPConfig contains configuration for Streamable HTTP transport.
type StreamableHTTPConfig struct {
	Addr              string
	EndpointPath      string
	HeartbeatInterval time.Duration
	Stateless         bool
}

// Server represents the MCP server instance.
type Server struct {
	mcpServer             *server.MCPServer
	transport             TransportType
	sseConfig             *SSEConfig
	streamableHTTPConfig  *StreamableHTTPConfig
}

// NewServer creates a new MCP server instance with default stdio transport.
func NewServer(name string, version string) (*Server, error) {
	// Create MCP server.
	mcpServer := server.NewMCPServer(
		name,
		version,
	)

	return &Server{
		mcpServer: mcpServer,
		transport: TransportStdio,
	}, nil
}

// WithTransport sets the transport type for the server.
func (s *Server) WithTransport(transport TransportType) *Server {
	s.transport = transport
	return s
}

// WithSSEConfig sets the SSE configuration for the server.
func (s *Server) WithSSEConfig(config *SSEConfig) *Server {
	s.sseConfig = config
	return s
}

// WithStreamableHTTPConfig sets the Streamable HTTP configuration for the server.
func (s *Server) WithStreamableHTTPConfig(config *StreamableHTTPConfig) *Server {
	s.streamableHTTPConfig = config
	return s
}

// RegisterTool registers a tool with the MCP server.
func (s *Server) RegisterTool(tool mcp.Tool, handler server.ToolHandlerFunc) error {
	s.mcpServer.AddTool(tool, handler)
	return nil
}

// Start starts the MCP server with the configured transport.
func (s *Server) Start(ctx context.Context) error {
	switch s.transport {
	case TransportStdio:
		return s.startStdio(ctx)
	case TransportSSE:
		return s.startSSE(ctx)
	case TransportStreamableHTTP:
		return s.startStreamableHTTP(ctx)
	default:
		return fmt.Errorf("unsupported transport type: %s", s.transport)
	}
}

// startStdio starts the server using stdio transport.
func (s *Server) startStdio(ctx context.Context) error {
	return server.ServeStdio(s.mcpServer)
}

// startSSE starts the server using SSE transport.
func (s *Server) startSSE(ctx context.Context) error {
	if s.sseConfig == nil {
		return fmt.Errorf("SSE configuration is required for SSE transport")
	}

	// Create SSE server options
	opts := []server.SSEOption{}
	if s.sseConfig.BaseURL != "" {
		opts = append(opts, server.WithBaseURL(s.sseConfig.BaseURL))
	}
	if s.sseConfig.StaticBasePath != "" {
		opts = append(opts, server.WithStaticBasePath(s.sseConfig.StaticBasePath))
	}
	if s.sseConfig.KeepAliveInterval > 0 {
		opts = append(opts, server.WithKeepAliveInterval(s.sseConfig.KeepAliveInterval))
	}

	// Create and start SSE server
	sseServer := server.NewSSEServer(s.mcpServer, opts...)
	return sseServer.Start(s.sseConfig.Addr)
}

// startStreamableHTTP starts the server using Streamable HTTP transport.
func (s *Server) startStreamableHTTP(ctx context.Context) error {
	if s.streamableHTTPConfig == nil {
		return fmt.Errorf("Streamable HTTP configuration is required for Streamable HTTP transport")
	}

	// Create Streamable HTTP server options
	opts := []server.StreamableHTTPOption{}
	if s.streamableHTTPConfig.EndpointPath != "" {
		opts = append(opts, server.WithEndpointPath(s.streamableHTTPConfig.EndpointPath))
	}
	if s.streamableHTTPConfig.HeartbeatInterval > 0 {
		opts = append(opts, server.WithHeartbeatInterval(s.streamableHTTPConfig.HeartbeatInterval))
	}
	if s.streamableHTTPConfig.Stateless {
		opts = append(opts, server.WithStateLess(true))
	}

	// Create and start Streamable HTTP server
	streamableServer := server.NewStreamableHTTPServer(s.mcpServer, opts...)
	return streamableServer.Start(s.streamableHTTPConfig.Addr)
}

// GetMCPServer returns the underlying MCP server instance.
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}
