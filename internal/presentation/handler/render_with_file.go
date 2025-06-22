package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/usecase"
)

// RenderWithFileHandler handles the d2_render tool with file output.
type RenderWithFileHandler struct {
	useCase *usecase.DiagramUseCase
}

// NewRenderWithFileHandler creates a new render handler.
func NewRenderWithFileHandler(useCase *usecase.DiagramUseCase) *RenderWithFileHandler {
	return &RenderWithFileHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *RenderWithFileHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_render_to_file",
		mcp.WithDescription("Render D2 text into a diagram and save to file"),
		mcp.WithString("content", mcp.Description("D2 diagram text content"), mcp.Required()),
		mcp.WithString("format", mcp.Description("Output format (svg, png, pdf)"), mcp.Enum("svg", "png", "pdf"), mcp.DefaultString("svg")),
		mcp.WithNumber("theme", mcp.Description("Theme ID (0-300+)"), mcp.DefaultNumber(0)),
	)
}

// GetHandler returns the tool handler function.
func (h *RenderWithFileHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the render request.
func (h *RenderWithFileHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Create output directory.
	outputDir := filepath.Join(os.TempDir(), "d2mcp_output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to create output directory", err), nil
	}

	// Generate filename.
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("diagram_%d.%s", timestamp, formatStr)
	filePath := filepath.Join(outputDir, filename)

	// Write to file.
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to write output file", err), nil
	}

	// Return file path and preview.
	result := fmt.Sprintf("Diagram saved to: %s\n\n", filePath)

	if format == entity.FormatSVG {
		// For SVG, also include a preview of the content
		result += fmt.Sprintf("SVG Preview (first 500 chars):\n%s...", string(data[:min(500, len(data))]))
	} else {
		// For binary formats, provide file info
		result += fmt.Sprintf("File size: %d bytes", len(data))
	}

	return mcp.NewToolResultText(result), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
