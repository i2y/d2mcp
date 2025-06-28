package handler

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// OracleDeleteHandler handles the d2_oracle_delete tool.
type OracleDeleteHandler struct {
	useCase *usecase.OracleUseCase
}

// NewOracleDeleteHandler creates a new Oracle delete handler.
func NewOracleDeleteHandler(useCase *usecase.OracleUseCase) *OracleDeleteHandler {
	return &OracleDeleteHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *OracleDeleteHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_oracle_delete",
		mcp.WithDescription("Remove shapes or connections from a D2 diagram. Use this when you need to: clean up unwanted elements, refactor diagram structure, or remove outdated components. Important: deleting a container shape will also delete ALL its child elements. Connections to/from deleted shapes are automatically removed. Use this carefully - consider using d2_oracle_move to relocate elements instead if you want to preserve them. Perfect for iterative diagram refinement and cleanup operations."),
		mcp.WithString("diagram_id", mcp.Description("ID of the diagram to modify"), mcp.Required()),
		mcp.WithString("key", mcp.Description("Key of the element to delete. Examples: 'server' for a shape, 'server -> database' for a connection, 'System.Database' for nested element. WARNING: Deleting containers removes all children"), mcp.Required()),
	)
}

// GetHandler returns the tool handler function.
func (h *OracleDeleteHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the delete request.
func (h *OracleDeleteHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagram_id", "")
	key := mcp.ParseString(request, "key", "")

	op := &entity.OracleOperation{
		Type:      entity.OracleDelete,
		DiagramID: diagramID,
		Key:       key,
		BoardPath: []string{}, // For now, single board support
	}

	result, err := h.useCase.DeleteElement(ctx, op)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to delete element", err), nil
	}

	response := fmt.Sprintf("Element '%s' deleted successfully", key)
	if len(result.IDDeltas) > 0 {
		response += fmt.Sprintf(" (affected %d related elements)", len(result.IDDeltas))
	}

	return mcp.NewToolResultText(response), nil
}
