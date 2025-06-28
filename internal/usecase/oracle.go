package usecase

import (
	"context"
	"fmt"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/domain/repository"
)

// OracleUseCase implements business logic for Oracle operations
type OracleUseCase struct {
	repo repository.OracleRepository
}

// NewOracleUseCase creates a new Oracle use case
func NewOracleUseCase(repo repository.OracleRepository) *OracleUseCase {
	return &OracleUseCase{repo: repo}
}

// CreateElement creates a new diagram element
func (uc *OracleUseCase) CreateElement(ctx context.Context, op *entity.OracleOperation) (*entity.OracleResult, error) {
	if op.DiagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}
	if op.Key == "" {
		return nil, &ValidationError{Message: "element key is required"}
	}

	return uc.repo.CreateElement(ctx, op.DiagramID, op.BoardPath, op.Key)
}

// SetAttribute sets an attribute on a diagram element
func (uc *OracleUseCase) SetAttribute(ctx context.Context, op *entity.OracleOperation) (*entity.OracleResult, error) {
	if op.DiagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}
	if op.Key == "" {
		return nil, &ValidationError{Message: "element key is required"}
	}

	return uc.repo.SetAttribute(ctx, op.DiagramID, op.BoardPath, op.Key, op.Tag, op.Value)
}

// DeleteElement deletes a diagram element
func (uc *OracleUseCase) DeleteElement(ctx context.Context, op *entity.OracleOperation) (*entity.OracleResult, error) {
	if op.DiagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}
	if op.Key == "" {
		return nil, &ValidationError{Message: "element key is required"}
	}

	return uc.repo.DeleteElement(ctx, op.DiagramID, op.BoardPath, op.Key)
}

// MoveElement moves a diagram element to a new container
func (uc *OracleUseCase) MoveElement(ctx context.Context, op *entity.OracleOperation) (*entity.OracleResult, error) {
	if op.DiagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}
	if op.Key == "" {
		return nil, &ValidationError{Message: "element key is required"}
	}
	if op.NewKey == nil || *op.NewKey == "" {
		return nil, &ValidationError{Message: "new key is required"}
	}

	return uc.repo.MoveElement(ctx, op.DiagramID, op.BoardPath, op.Key, *op.NewKey, op.IncludeDescendants)
}

// RenameElement renames a diagram element
func (uc *OracleUseCase) RenameElement(ctx context.Context, op *entity.OracleOperation) (*entity.OracleResult, error) {
	if op.DiagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}
	if op.Key == "" {
		return nil, &ValidationError{Message: "element key is required"}
	}
	if op.NewKey == nil || *op.NewKey == "" {
		return nil, &ValidationError{Message: "new name is required"}
	}

	return uc.repo.RenameElement(ctx, op.DiagramID, op.BoardPath, op.Key, *op.NewKey)
}

// GetObject retrieves object information
func (uc *OracleUseCase) GetObject(ctx context.Context, diagramID string, boardPath []string, objectID string) (*entity.GraphObject, error) {
	if diagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}
	if objectID == "" {
		return nil, &ValidationError{Message: "object ID is required"}
	}

	return uc.repo.GetObject(ctx, diagramID, boardPath, objectID)
}

// GetEdge retrieves edge information
func (uc *OracleUseCase) GetEdge(ctx context.Context, diagramID string, boardPath []string, edgeID string) (*entity.GraphEdge, error) {
	if diagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}
	if edgeID == "" {
		return nil, &ValidationError{Message: "edge ID is required"}
	}

	return uc.repo.GetEdge(ctx, diagramID, boardPath, edgeID)
}

// GetChildren retrieves child element IDs
func (uc *OracleUseCase) GetChildren(ctx context.Context, diagramID string, boardPath []string, parentID string) ([]string, error) {
	if diagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}

	return uc.repo.GetChildren(ctx, diagramID, boardPath, parentID)
}

// LoadDiagram loads a diagram from D2 text
func (uc *OracleUseCase) LoadDiagram(ctx context.Context, diagramID string, content string) error {
	if diagramID == "" {
		return &ValidationError{Message: "diagram ID is required"}
	}
	if content == "" {
		return &ValidationError{Message: "diagram content is required"}
	}

	return uc.repo.LoadDiagram(ctx, diagramID, content)
}

// SerializeDiagram converts the current graph state back to D2 text
func (uc *OracleUseCase) SerializeDiagram(ctx context.Context, diagramID string) (string, error) {
	if diagramID == "" {
		return "", &ValidationError{Message: "diagram ID is required"}
	}

	return uc.repo.SerializeDiagram(ctx, diagramID)
}

// ExecuteOperation executes a single Oracle operation based on its type
func (uc *OracleUseCase) ExecuteOperation(ctx context.Context, op *entity.OracleOperation) (*entity.OracleResult, error) {
	switch op.Type {
	case entity.OracleCreate:
		return uc.CreateElement(ctx, op)
	case entity.OracleSet:
		return uc.SetAttribute(ctx, op)
	case entity.OracleDelete:
		return uc.DeleteElement(ctx, op)
	case entity.OracleMove:
		return uc.MoveElement(ctx, op)
	case entity.OracleRename:
		return uc.RenameElement(ctx, op)
	default:
		return nil, fmt.Errorf("unknown operation type: %s", op.Type)
	}
}
