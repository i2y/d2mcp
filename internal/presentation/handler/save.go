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

// SaveHandler handles the d2_save tool.
type SaveHandler struct {
	useCase *usecase.DiagramUseCase
}

// NewSaveHandler creates a new save handler.
func NewSaveHandler(useCase *usecase.DiagramUseCase) *SaveHandler {
	return &SaveHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *SaveHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_save",
		mcp.WithDescription("Save an existing diagram to a file on disk. The diagram must be created first using d2_create. This tool exports the diagram in the specified format and writes it to a file path, returning the path where it was saved. Supported formats: svg (default), png, pdf. If no path is provided, saves to a temporary directory with a timestamped filename. Path handling: absolute paths (e.g., /Users/name/diagram.svg) are used as-is; relative paths (e.g., diagram.svg) are resolved from the MCP server's working directory. When unsure, either use absolute paths or omit the path to use the temp directory."),
		mcp.WithString("diagramId", mcp.Description("ID of the diagram to save"), mcp.Required()),
		mcp.WithString("format", mcp.Description("Export format (svg, png, pdf)"), mcp.Enum("svg", "png", "pdf"), mcp.DefaultString("svg")),
		mcp.WithString("path", mcp.Description("Output file path. Examples: '/Users/name/diagram.svg' (absolute), 'output/diagram.svg' (relative to MCP server), or omit for auto-generated path in temp directory")),
	)
}

// GetHandler returns the tool handler function.
func (h *SaveHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the save request.
func (h *SaveHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagramId", "")
	if diagramID == "" {
		return mcp.NewToolResultError("diagramId is required"), nil
	}

	formatStr := mcp.ParseString(request, "format", "svg")
	format := entity.ExportFormat(formatStr)

	outputPath := mcp.ParseString(request, "path", "")

	// Export the diagram.
	reader, err := h.useCase.ExportDiagram(ctx, diagramID, format)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to export diagram", err), nil
	}

	// Read the output.
	data, err := io.ReadAll(reader)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to read exported output", err), nil
	}

	// Determine output path.
	if outputPath == "" {
		// Create output directory in temp.
		outputDir := filepath.Join(os.TempDir(), "d2mcp_output")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to create output directory", err), nil
		}

		// Generate filename.
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("%s_%d.%s", diagramID, timestamp, formatStr)
		outputPath = filepath.Join(outputDir, filename)
	} else {
		// Ensure directory exists.
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to create output directory", err), nil
		}
	}

	// Write to file.
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to write output file", err), nil
	}

	// Return result.
	result := fmt.Sprintf("Diagram saved to: %s\n", outputPath)
	result += fmt.Sprintf("Format: %s\n", formatStr)
	result += fmt.Sprintf("Size: %d bytes", len(data))

	return mcp.NewToolResultText(result), nil
}
