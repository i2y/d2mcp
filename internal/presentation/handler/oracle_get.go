package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/i2y/d2mcp/internal/usecase"
)

// OracleGetHandler handles the d2_oracle_get_info tool.
type OracleGetHandler struct {
	useCase *usecase.OracleUseCase
}

// NewOracleGetHandler creates a new Oracle get handler.
func NewOracleGetHandler(useCase *usecase.OracleUseCase) *OracleGetHandler {
	return &OracleGetHandler{
		useCase: useCase,
	}
}

// GetTool returns the MCP tool definition.
func (h *OracleGetHandler) GetTool() mcp.Tool {
	return mcp.NewTool(
		"d2_oracle_get_info",
		mcp.WithDescription("Inspect and analyze diagram elements to understand structure and properties. Use this when you need to: verify element exists before modifying, check current properties/attributes, explore container contents, debug connection issues, or understand diagram hierarchy. Info types: 'object' returns shape details (labels, styles, attributes), 'edge' returns connection properties (labels, arrows, styles), 'children' lists all elements inside a container. Essential for safe modifications - always check before changing. Returns JSON with complete element information."),
		mcp.WithString("diagram_id", mcp.Description("ID of the diagram"), mcp.Required()),
		mcp.WithString("key", mcp.Description("Key of the element to inspect. Examples: 'server' for shape info, 'server -> database' for connection info, 'System' to see what's inside a container"), mcp.Required()),
		mcp.WithString("info_type", mcp.Description("Type of information to retrieve: 'object' for shape/container details, 'edge' for connection properties, 'children' to list elements inside a container"), mcp.DefaultString("object")),
	)
}

// GetHandler returns the tool handler function.
func (h *OracleGetHandler) GetHandler() server.ToolHandlerFunc {
	return h.Handle
}

// Handle processes the get request.
func (h *OracleGetHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments.
	diagramID := mcp.ParseString(request, "diagram_id", "")
	key := mcp.ParseString(request, "key", "")
	infoType := mcp.ParseString(request, "info_type", "object")

	boardPath := []string{} // For now, single board support

	switch infoType {
	case "object":
		obj, err := h.useCase.GetObject(ctx, diagramID, boardPath, key)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to get object info", err), nil
		}

		jsonData, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("Failed to format object info"), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Object info for '%s':\n%s", key, string(jsonData))), nil

	case "edge":
		edge, err := h.useCase.GetEdge(ctx, diagramID, boardPath, key)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to get edge info", err), nil
		}

		jsonData, err := json.MarshalIndent(edge, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("Failed to format edge info"), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Edge info for '%s':\n%s", key, string(jsonData))), nil

	case "children":
		children, err := h.useCase.GetChildren(ctx, diagramID, boardPath, key)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to get children", err), nil
		}

		if len(children) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No children found for '%s'", key)), nil
		}

		jsonData, err := json.MarshalIndent(children, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("Failed to format children info"), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Children of '%s':\n%s", key, string(jsonData))), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf("Invalid info_type: %s. Must be 'object', 'edge', or 'children'", infoType)), nil
	}
}
