package handler

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// OracleCreateHandler handles the d2_oracle_create tool.
type OracleCreateHandler struct {
	useCase *usecase.OracleUseCase
}

// NewOracleCreateHandler creates a new Oracle create handler.
func NewOracleCreateHandler(useCase *usecase.OracleUseCase) *OracleCreateHandler {
	return &OracleCreateHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *OracleCreateHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_oracle_create",
		mcp.WithDescription("Create a new shape or connection in a D2 diagram"),
		mcp.WithString("diagram_id", mcp.Description("ID of the diagram to modify"), mcp.Required()),
		mcp.WithString("key", mcp.Description("Key for the new element (e.g., 'server' or 'server -> database')"), mcp.Required()),
	)
}

// GetHandler returns the tool handler function.
func (h *OracleCreateHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the create request.
func (h *OracleCreateHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagram_id", "")
	key := mcp.ParseString(request, "key", "")

	op := &entity.OracleOperation{
		Type:      entity.OracleCreate,
		DiagramID: diagramID,
		Key:       key,
		BoardPath: []string{}, // For now, single board support
	}

	result, err := h.useCase.CreateElement(ctx, op)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to create element", err), nil
	}

	response := fmt.Sprintf("Element created successfully. New key: %s", result.NewKey)
	if result.NewKey != key {
		response += fmt.Sprintf(" (auto-generated from '%s')", key)
	}

	return mcp.NewToolResultText(response), nil
}
