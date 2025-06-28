package repository

import (
	"context"

	"github.com/i2y/d2mcp/internal/domain/entity"
)

// OracleRepository defines Oracle API operations for incremental diagram manipulation
type OracleRepository interface {
	DiagramRepository // Embed existing interface

	// CreateElement creates a new shape or connection
	CreateElement(ctx context.Context, diagramID string, boardPath []string, key string) (*entity.OracleResult, error)

	// SetAttribute sets attributes on a shape or connection
	SetAttribute(ctx context.Context, diagramID string, boardPath []string, key string, tag, value *string) (*entity.OracleResult, error)

	// DeleteElement deletes a shape or connection
	DeleteElement(ctx context.Context, diagramID string, boardPath []string, key string) (*entity.OracleResult, error)

	// MoveElement moves a shape to a new container
	MoveElement(ctx context.Context, diagramID string, boardPath []string, key, newKey string, includeDescendants bool) (*entity.OracleResult, error)

	// RenameElement renames a shape or connection
	RenameElement(ctx context.Context, diagramID string, boardPath []string, key, newName string) (*entity.OracleResult, error)

	// GetObject retrieves object information
	GetObject(ctx context.Context, diagramID string, boardPath []string, objectID string) (*entity.GraphObject, error)

	// GetEdge retrieves edge information
	GetEdge(ctx context.Context, diagramID string, boardPath []string, edgeID string) (*entity.GraphEdge, error)

	// GetChildren retrieves child element IDs
	GetChildren(ctx context.Context, diagramID string, boardPath []string, parentID string) ([]string, error)

	// LoadDiagram loads a diagram from D2 text
	LoadDiagram(ctx context.Context, diagramID string, content string) error

	// SerializeDiagram converts the current graph state back to D2 text
	SerializeDiagram(ctx context.Context, diagramID string) (string, error)
}
