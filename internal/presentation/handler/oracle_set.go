package handler

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// OracleSetHandler handles the d2_oracle_set tool.
type OracleSetHandler struct {
	useCase *usecase.OracleUseCase
}

// NewOracleSetHandler creates a new Oracle set handler.
func NewOracleSetHandler(useCase *usecase.OracleUseCase) *OracleSetHandler {
	return &OracleSetHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *OracleSetHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_oracle_set",
		mcp.WithDescription("Set attributes on a shape or connection in a D2 diagram"),
		mcp.WithString("diagram_id", mcp.Description("ID of the diagram to modify"), mcp.Required()),
		mcp.WithString("key", mcp.Description("Key of the element to modify (e.g., 'server.shape' or 'connection.style.stroke')"), mcp.Required()),
		mcp.WithString("value", mcp.Description("The value to set"), mcp.Required()),
		mcp.WithString("tag", mcp.Description("Optional tag for the attribute (e.g., 'label' or 'style')")),
	)
}

// GetHandler returns the tool handler function.
func (h *OracleSetHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the set request.
func (h *OracleSetHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagram_id", "")
	key := mcp.ParseString(request, "key", "")
	value := mcp.ParseString(request, "value", "")
	tag := mcp.ParseString(request, "tag", "")

	var tagPtr, valuePtr *string
	if value != "" {
		valuePtr = &value
	}
	if tag != "" {
		tagPtr = &tag
	}

	op := &entity.OracleOperation{
		Type:      entity.OracleSet,
		DiagramID: diagramID,
		Key:       key,
		Value:     valuePtr,
		Tag:       tagPtr,
		BoardPath: []string{}, // For now, single board support
	}

	_, err := h.useCase.SetAttribute(ctx, op)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to set attribute", err), nil
	}

	response := fmt.Sprintf("Attribute set successfully: %s = %s", key, value)
	if tag != "" {
		response = fmt.Sprintf("Attribute set successfully: %s.%s = %s", key, tag, value)
	}

	return mcp.NewToolResultText(response), nil
}
