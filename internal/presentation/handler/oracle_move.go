package handler

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// OracleMoveHandler handles the d2_oracle_move tool.
type OracleMoveHandler struct {
	useCase *usecase.OracleUseCase
}

// NewOracleMoveHandler creates a new Oracle move handler.
func NewOracleMoveHandler(useCase *usecase.OracleUseCase) *OracleMoveHandler {
	return &OracleMoveHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *OracleMoveHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_oracle_move",
		mcp.WithDescription("Reorganize diagram structure by moving shapes between containers. Use this when you need to: group related components together, refactor diagram hierarchy, move elements into or out of systems/packages, or restructure without losing connections. Containers are shapes that hold other shapes (like 'System', 'Network', or any shape with children). Moving preserves all connections - they're automatically rerouted. Set include_descendants=false to move only the parent shape, leaving children in original location. Essential for maintaining clean, logical diagram organization."),
		mcp.WithString("diagram_id", mcp.Description("ID of the diagram to modify"), mcp.Required()),
		mcp.WithString("key", mcp.Description("Key of the element to move (e.g., 'server', 'Database.users_table')"), mcp.Required()),
		mcp.WithString("new_parent", mcp.Description("Target container key where element will be moved. Use empty string '' to move to root level. Examples: 'System' to move into System container, 'Network.DMZ' for nested container"), mcp.Required()),
		mcp.WithString("include_descendants", mcp.Description("Whether to move child elements along with the parent (true/false). Default true preserves hierarchy"), mcp.DefaultString("true")),
	)
}

// GetHandler returns the tool handler function.
func (h *OracleMoveHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the move request.
func (h *OracleMoveHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagram_id", "")
	key := mcp.ParseString(request, "key", "")
	newParent := mcp.ParseString(request, "new_parent", "")
	includeDescendantsStr := mcp.ParseString(request, "include_descendants", "true")
	includeDescendants := includeDescendantsStr == "true"

	// Build new key path
	newKey := newParent
	if newParent != "" {
		newKey = newParent + "." + key
	}

	op := &entity.OracleOperation{
		Type:               entity.OracleMove,
		DiagramID:          diagramID,
		Key:                key,
		NewKey:             &newKey,
		IncludeDescendants: includeDescendants,
		BoardPath:          []string{}, // For now, single board support
	}

	_, err := h.useCase.MoveElement(ctx, op)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to move element", err), nil
	}

	response := fmt.Sprintf("Element '%s' moved to '%s'", key, newKey)
	if includeDescendants {
		response += " (including descendants)"
	}

	return mcp.NewToolResultText(response), nil
}
