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
		mcp.WithDescription("Create a new diagram that can be edited with Oracle API tools. This is the unified way to create diagrams:\n\n1. Empty diagram (no content): For building incrementally with Oracle API\n2. From D2 text (with content): For rendering complete D2 diagrams\n\nBoth types are fully editable using d2_oracle_* tools.\n\nExamples:\n- d2_create(id=\"arch\") → Empty diagram for incremental building\n- d2_create(id=\"arch\", content=\"a -> b\") → Diagram from D2 text\n\nUse cases:\n- Building diagrams from data sources (use empty)\n- Rendering complete D2 text (use with content)\n- Converting existing D2 to editable form (use with content)\n- Interactive diagram creation (use empty)"),
		mcp.WithString("id", mcp.Description("Unique identifier for the diagram"), mcp.Required()),
		mcp.WithString("content", mcp.Description("Optional D2 text content. If provided, creates a diagram from this content (which can then be edited with Oracle API). If not provided, creates an empty diagram for incremental building. Both are fully editable."), mcp.DefaultString("")),
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

	// Provide appropriate feedback based on whether content was provided
	var message string
	if content != "" {
		message = fmt.Sprintf("Diagram '%s' created successfully from provided D2 content. You can now use d2_oracle_* tools to modify it, or d2_export to render it.", id)
	} else {
		message = fmt.Sprintf("Empty diagram '%s' created successfully. Use d2_oracle_create to add shapes and connections.", id)
	}

	return mcp.NewToolResultText(message), nil
}
