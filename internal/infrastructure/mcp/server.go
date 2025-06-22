package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the MCP server instance.
type Server struct {
	mcpServer *server.MCPServer
}

// NewServer creates a new MCP server instance.
func NewServer(name string, version string) (*Server, error) {
	// Create MCP server.
	mcpServer := server.NewMCPServer(
		name,
		version,
	)

	return &Server{
		mcpServer: mcpServer,
	}, nil
}

// RegisterTool registers a tool with the MCP server.
func (s *Server) RegisterTool(tool mcp.Tool, handler server.ToolHandlerFunc) error {
	s.mcpServer.AddTool(tool, handler)
	return nil
}

// Start starts the MCP server.
func (s *Server) Start(ctx context.Context) error {
	// Use ServeStdio directly.
	return server.ServeStdio(s.mcpServer)
}

// GetMCPServer returns the underlying MCP server instance.
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}
