package d2

import (
	"context"
	"fmt"
	"testing"

	"github.com/i2y/d2mcp/internal/domain/entity"
)

func TestD2Repository_Render(t *testing.T) {
	repo := NewD2Repository()
	ctx := context.Background()

	tests := []struct {
		name    string
		content string
		format  entity.ExportFormat
		wantErr bool
	}{
		{
			name:    "render SVG",
			content: "a -> b",
			format:  entity.FormatSVG,
			wantErr: false,
		},
		{
			name:    "render PNG requires external tool",
			content: "a -> b",
			format:  entity.FormatPNG,
			wantErr: true, // Will fail without rsvg-convert or imagemagick
		},
		{
			name:    "render PDF requires external tool",
			content: "a -> b",
			format:  entity.FormatPDF,
			wantErr: true, // Will fail without rsvg-convert or imagemagick
		},
		{
			name:    "invalid content",
			content: "invalid -> -> syntax",
			format:  entity.FormatSVG,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := repo.Render(ctx, tt.content, tt.format, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && reader == nil {
				t.Error("Render() returned nil reader without error")
			}
		})
	}
}

func TestD2Repository_CreateAndExport(t *testing.T) {
	repo := NewD2Repository()
	ctx := context.Background()

	// Test Create
	diagram := &entity.Diagram{
		ID:      "test-diagram",
		Content: "a -> b",
	}

	if err := repo.Create(ctx, diagram); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Test Export
	reader, err := repo.Export(ctx, diagram.ID, entity.FormatSVG)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if reader == nil {
		t.Error("Export() returned nil reader")
	}

	// Test Export non-existent diagram
	_, err = repo.Export(ctx, "non-existent", entity.FormatSVG)
	if err == nil {
		t.Error("Export() should fail for non-existent diagram")
	}
}

func TestD2Repository_ConcurrentAccess(t *testing.T) {
	repo := NewD2Repository()
	ctx := context.Background()

	// Create multiple diagrams
	for i := 0; i < 5; i++ {
		diagram := &entity.Diagram{
			ID:      fmt.Sprintf("concurrent-test-%d", i),
			Content: fmt.Sprintf("a%d -> b%d", i, i),
		}

		if err := repo.Create(ctx, diagram); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Run concurrent operations
	done := make(chan bool)
	errors := make(chan error, 10)

	// Concurrent reads/exports
	for i := 0; i < 5; i++ {
		go func(i int) {
			_, err := repo.Export(ctx, fmt.Sprintf("concurrent-test-%d", i), entity.FormatSVG)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Concurrent renders
	for i := 0; i < 5; i++ {
		go func(i int) {
			_, err := repo.Render(ctx, fmt.Sprintf("x%d -> y%d", i, i), entity.FormatSVG, nil)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent operation failed: %v", err)
	default:
		// No errors
	}
}
