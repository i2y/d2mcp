package handler

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// CreateHandler handles the d2_create tool.
type CreateHandler struct {
	useCase *usecase.DiagramUseCase
}

// NewCreateHandler creates a new create handler.
func NewCreateHandler(useCase *usecase.DiagramUseCase) *CreateHandler {
	return &CreateHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *CreateHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_create",
		mcp.WithDescription("Create a new D2 diagram programmatically"),
		mcp.WithString("id", mcp.Description("Unique identifier for the diagram"), mcp.Required()),
		mcp.WithString("content", mcp.Description("Initial D2 content (optional)"), mcp.DefaultString("")),
	)
}

// GetHandler returns the tool handler function.
func (h *CreateHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the create request.
func (h *CreateHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	id := mcp.ParseString(request, "id", "")
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	content := mcp.ParseString(request, "content", "")

	// Create the diagram.
	diagram := &entity.Diagram{
		ID:      id,
		Content: content,
	}

	err := h.useCase.CreateDiagram(ctx, diagram)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to create diagram", err), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Diagram '%s' created successfully", id)), nil
}
