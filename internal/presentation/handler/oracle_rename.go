package handler

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// OracleRenameHandler handles the d2_oracle_rename tool.
type OracleRenameHandler struct {
	useCase *usecase.OracleUseCase
}

// NewOracleRenameHandler creates a new Oracle rename handler.
func NewOracleRenameHandler(useCase *usecase.OracleUseCase) *OracleRenameHandler {
	return &OracleRenameHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *OracleRenameHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_oracle_rename",
		mcp.WithDescription("Change the identifier of shapes or connections while preserving all relationships. Use this when you need to: improve clarity with better names, fix typos or naming inconsistencies, refactor diagram elements, or align with updated terminology. The rename is intelligent - ALL connections referencing the old name are automatically updated to use the new name. This includes connections where the element is source, target, or part of a longer path. Child elements keep their relative names. Safe operation that maintains diagram integrity."),
		mcp.WithString("diagram_id", mcp.Description("ID of the diagram to modify"), mcp.Required()),
		mcp.WithString("key", mcp.Description("Current key of the element to rename (e.g., 'server', 'DB', 'System.OldName')"), mcp.Required()),
		mcp.WithString("new_name", mcp.Description("New identifier for the element (e.g., 'web_server', 'Database', 'NewName'). Connections are automatically updated"), mcp.Required()),
	)
}

// GetHandler returns the tool handler function.
func (h *OracleRenameHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the rename request.
func (h *OracleRenameHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagram_id", "")
	key := mcp.ParseString(request, "key", "")
	newName := mcp.ParseString(request, "new_name", "")

	op := &entity.OracleOperation{
		Type:      entity.OracleRename,
		DiagramID: diagramID,
		Key:       key,
		NewKey:    &newName,
		BoardPath: []string{}, // For now, single board support
	}

	result, err := h.useCase.RenameElement(ctx, op)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to rename element", err), nil
	}

	response := fmt.Sprintf("Element renamed from '%s' to '%s'", key, newName)
	if len(result.IDDeltas) > 0 {
		response += fmt.Sprintf(" (updated %d references)", len(result.IDDeltas))
	}

	return mcp.NewToolResultText(response), nil
}
