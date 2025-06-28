package usecase

import (
	"context"
	"io"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/domain/repository"
)

// DiagramUseCase implements the business logic for diagram operations.
type DiagramUseCase struct {
	repo repository.DiagramRepository
}

// NewDiagramUseCase creates a new diagram usecase instance.
func NewDiagramUseCase(repo repository.DiagramRepository) *DiagramUseCase {
	return &DiagramUseCase{
		repo: repo,
	}
}

// RenderDiagram renders D2 text into a diagram.
func (uc *DiagramUseCase) RenderDiagram(ctx context.Context, content string, format entity.ExportFormat, theme *entity.Theme) (io.Reader, error) {
	// Validate input.
	if content == "" {
		return nil, &ValidationError{Message: "content cannot be empty"}
	}

	// Default to SVG if no format specified.
	if format == "" {
		format = entity.FormatSVG
	}

	return uc.repo.Render(ctx, content, format, theme)
}

// CreateDiagram creates a new diagram programmatically.
func (uc *DiagramUseCase) CreateDiagram(ctx context.Context, diagram *entity.Diagram) error {
	// Validate diagram.
	if diagram == nil {
		return &ValidationError{Message: "diagram cannot be nil"}
	}
	if diagram.ID == "" {
		return &ValidationError{Message: "diagram ID is required"}
	}

	return uc.repo.Create(ctx, diagram)
}

// ExportDiagram exports the diagram to the specified format.
func (uc *DiagramUseCase) ExportDiagram(ctx context.Context, diagramID string, format entity.ExportFormat) (io.Reader, error) {
	// Validate input.
	if diagramID == "" {
		return nil, &ValidationError{Message: "diagram ID is required"}
	}

	// Default to SVG if no format specified.
	if format == "" {
		format = entity.FormatSVG
	}

	return uc.repo.Export(ctx, diagramID, format)
}

// Create creates a diagram with the given ID and optional content.
// This is a convenience method that handles both empty and pre-populated diagrams.
func (uc *DiagramUseCase) Create(ctx context.Context, id string, content string) error {
	diagram := &entity.Diagram{
		ID:      id,
		Content: content,
	}
	return uc.CreateDiagram(ctx, diagram)
}
