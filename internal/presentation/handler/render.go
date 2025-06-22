package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// RenderHandler handles the d2_render tool.
type RenderHandler struct {
	useCase *usecase.DiagramUseCase
}

// NewRenderHandler creates a new render handler.
func NewRenderHandler(useCase *usecase.DiagramUseCase) *RenderHandler {
	return &RenderHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *RenderHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_render",
		mcp.WithDescription("Render D2 text into a diagram"),
		mcp.WithString("content", mcp.Description("D2 diagram text content"), mcp.Required()),
		mcp.WithString("format", mcp.Description("Output format (svg, png, pdf)"), mcp.Enum("svg", "png", "pdf"), mcp.DefaultString("svg")),
		mcp.WithNumber("theme", mcp.Description("Theme ID (0-300+)"), mcp.DefaultNumber(0)),
	)
}

// GetHandler returns the tool handler function.
func (h *RenderHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the render request.
func (h *RenderHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	content := mcp.ParseString(request, "content", "")
	if content == "" {
		return mcp.NewToolResultError("content is required"), nil
	}

	formatStr := mcp.ParseString(request, "format", "svg")
	format := entity.ExportFormat(formatStr)

	var theme *entity.Theme
	themeID := mcp.ParseInt(request, "theme", 0)
	// Always create a theme object to pass the theme ID, even for ID 0
	theme = &entity.Theme{
		ID: themeID,
	}

	// Render the diagram.
	reader, err := h.useCase.RenderDiagram(ctx, content, format, theme)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to render diagram", err), nil
	}

	// Read the output.
	data, err := io.ReadAll(reader)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to read rendered output", err), nil
	}

	// Return result based on format.
	if format == entity.FormatSVG {
		// Return plain SVG - Claude Desktop should handle this correctly
		return mcp.NewToolResultText(string(data)), nil
	}

	// For binary formats, encode as base64.
	encoded := base64.StdEncoding.EncodeToString(data)
	return mcp.NewToolResultText(fmt.Sprintf("data:%s;base64,%s", getMimeType(format), encoded)), nil
}

// getMimeType returns the MIME type for the given format.
func getMimeType(format entity.ExportFormat) string {
	switch format {
	case entity.FormatPNG:
		return "image/png"
	case entity.FormatPDF:
		return "application/pdf"
	default:
		return "image/svg+xml"
	}
}
