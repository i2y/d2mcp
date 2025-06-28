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

// ExportHandler handles diagram export operations.
type ExportHandler struct {
	useCase *usecase.DiagramUseCase
}

// NewExportHandler creates a new export handler.
func NewExportHandler(useCase *usecase.DiagramUseCase) *ExportHandler {
	return &ExportHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *ExportHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_export",
		mcp.WithDescription("Export an existing diagram to SVG, PNG, or PDF format. The diagram must first be created using d2_create (not d2_render). Supports exporting all D2 features including SQL tables, UML classes, sequence diagrams, code blocks, and markdown-rich documentation. Note: PNG and PDF formats require external tools (e.g., Chromium) to be installed on the system."),
		mcp.WithString("diagramId", mcp.Description("ID of the diagram to export"), mcp.Required()),
		mcp.WithString("format", mcp.Description("Export format (svg, png, pdf)"), mcp.Enum("svg", "png", "pdf"), mcp.DefaultString("svg")),
	)
}

// GetHandler returns the tool handler function.
func (h *ExportHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the export request.
func (h *ExportHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagramId", "")
	if diagramID == "" {
		return mcp.NewToolResultError("diagramId is required"), nil
	}

	formatStr := mcp.ParseString(request, "format", "svg")
	format := entity.ExportFormat(formatStr)

	// Export the diagram.
	reader, err := h.useCase.ExportDiagram(ctx, diagramID, format)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to export diagram", err), nil
	}

	// Read the output.
	data, err := io.ReadAll(reader)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to read exported data", err), nil
	}

	// Return result based on format.
	if format == entity.FormatSVG {
		return mcp.NewToolResultText(string(data)), nil
	}

	// For binary formats, encode as base64.
	encoded := base64.StdEncoding.EncodeToString(data)
	mimeType := getMimeType(format)

	return mcp.NewToolResultText(fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)), nil
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
