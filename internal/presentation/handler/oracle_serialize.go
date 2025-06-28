package handler

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/usecase"
)

// OracleSerializeHandler handles the d2_oracle_serialize tool.
type OracleSerializeHandler struct {
	useCase *usecase.OracleUseCase
}

// NewOracleSerializeHandler creates a new Oracle serialize handler.
func NewOracleSerializeHandler(useCase *usecase.OracleUseCase) *OracleSerializeHandler {
	return &OracleSerializeHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *OracleSerializeHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_oracle_serialize",
		mcp.WithDescription("Get the current D2 text representation of a diagram being edited with Oracle API"),
		mcp.WithString("diagram_id", mcp.Description("ID of the diagram to serialize"), mcp.Required()),
	)
}

// GetHandler returns the tool handler function.
func (h *OracleSerializeHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the serialize request.
func (h *OracleSerializeHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagram_id", "")

	content, err := h.useCase.SerializeDiagram(ctx, diagramID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to serialize diagram", err), nil
	}

	return mcp.NewToolResultText(content), nil
}