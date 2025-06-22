package d2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/domain/repository"
)

// D2Repository implements the DiagramRepository interface using D2.
type D2Repository struct {
	diagrams map[string]*diagramData
	mu       sync.RWMutex
}

// diagramData holds the D2 graph and related data.
type diagramData struct {
	content string
	graph   *d2graph.Graph
}

// NewD2Repository creates a new D2 repository instance.
func NewD2Repository() repository.DiagramRepository {
	return &D2Repository{
		diagrams: make(map[string]*diagramData),
	}
}

// withSilentD2 executes a function with D2 logging disabled.
func withSilentD2(ctx context.Context, fn func(context.Context) error) error {
	// Create a null logger to prevent D2 from logging.
	nullLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx = log.With(ctx, nullLogger)

	// Temporarily redirect stderr to prevent any output.
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() {
		os.Stderr = oldStderr
	}()

	return fn(ctx)
}

// Render renders D2 text into a diagram with specified format.
func (r *D2Repository) Render(ctx context.Context, content string, format entity.ExportFormat, theme *entity.Theme) (io.Reader, error) {
	var result io.Reader
	err := withSilentD2(ctx, func(ctx context.Context) error {
		// Create ruler for text measurement.
		ruler, err := textmeasure.NewRuler()
		if err != nil {
			return fmt.Errorf("failed to create ruler: %w", err)
		}

		// Create layout resolver.
		layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
			return d2dagrelayout.DefaultLayout, nil
		}

		// Create compile options.
		compileOpts := &d2lib.CompileOptions{
			LayoutResolver: layoutResolver,
			Ruler:          ruler,
		}

		// Create render options.
		pad := int64(d2svg.DEFAULT_PADDING)
		renderOpts := &d2svg.RenderOpts{
			Pad: &pad,
		}

		// Apply theme if provided
		if theme != nil {
			themeID := int64(theme.ID)
			renderOpts.ThemeID = &themeID
		}

		// Compile the D2 script.
		diagram, _, err := d2lib.Compile(ctx, content, compileOpts, renderOpts)
		if err != nil {
			return fmt.Errorf("failed to compile D2 script: %w", err)
		}

		// Render based on format.
		switch format {
		case entity.FormatSVG, "":
			svg, err := d2svg.Render(diagram, renderOpts)
			if err != nil {
				return fmt.Errorf("failed to render SVG: %w", err)
			}
			result = bytes.NewReader(svg)
			return nil

		case entity.FormatPNG:
			// First render as SVG
			svg, err := d2svg.Render(diagram, renderOpts)
			if err != nil {
				return fmt.Errorf("failed to render SVG: %w", err)
			}

			// Note: tdewolff/canvas has issues parsing complex D2-generated SVGs
			// with embedded styles and fonts. Falling back to external tools.
			// TODO: Consider alternative pure-Go SVG rasterizers or preprocessing SVG

			// Create temporary files
			tmpDir := os.TempDir()
			svgFile, err := os.CreateTemp(tmpDir, "d2-*.svg")
			if err != nil {
				return fmt.Errorf("failed to create temp SVG file: %w", err)
			}
			defer os.Remove(svgFile.Name())

			// Write SVG to temp file
			if _, err := svgFile.Write(svg); err != nil {
				svgFile.Close()
				return fmt.Errorf("failed to write SVG: %w", err)
			}
			svgFile.Close()

			// Prepare output file
			outFile, err := os.CreateTemp(tmpDir, "d2-*.png")
			if err != nil {
				return fmt.Errorf("failed to create temp output file: %w", err)
			}
			defer os.Remove(outFile.Name())
			outFile.Close()

			// Try to convert using rsvg-convert (common on Unix systems)
			cmd := exec.Command("rsvg-convert", "-f", "png", "-o", outFile.Name(), svgFile.Name())

			if err := cmd.Run(); err != nil {
				// Try ImageMagick convert as fallback
				cmd = exec.Command("convert", svgFile.Name(), outFile.Name())
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to convert SVG to PNG: %w (install rsvg-convert or imagemagick)", err)
				}
			}

			// Read the converted file
			output, err := os.ReadFile(outFile.Name())
			if err != nil {
				return fmt.Errorf("failed to read converted file: %w", err)
			}

			result = bytes.NewReader(output)
			return nil

		case entity.FormatPDF:
			// First render as SVG
			svg, err := d2svg.Render(diagram, renderOpts)
			if err != nil {
				return fmt.Errorf("failed to render SVG: %w", err)
			}

			// For PDF, still use external tools for now
			// TODO: Consider using canvas for PDF rendering as well
			tmpDir := os.TempDir()
			svgFile, err := os.CreateTemp(tmpDir, "d2-*.svg")
			if err != nil {
				return fmt.Errorf("failed to create temp SVG file: %w", err)
			}
			defer os.Remove(svgFile.Name())

			// Write SVG to temp file
			if _, err := svgFile.Write(svg); err != nil {
				svgFile.Close()
				return fmt.Errorf("failed to write SVG: %w", err)
			}
			svgFile.Close()

			// Prepare output file
			outFile, err := os.CreateTemp(tmpDir, "d2-*.pdf")
			if err != nil {
				return fmt.Errorf("failed to create temp output file: %w", err)
			}
			defer os.Remove(outFile.Name())
			outFile.Close()

			// Try to convert using rsvg-convert (common on Unix systems)
			cmd := exec.Command("rsvg-convert", "-f", "pdf", "-o", outFile.Name(), svgFile.Name())

			if err := cmd.Run(); err != nil {
				// Try ImageMagick convert as fallback
				cmd = exec.Command("convert", svgFile.Name(), outFile.Name())
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to convert SVG to PDF: %w (install rsvg-convert or imagemagick)", err)
				}
			}

			// Read the converted file
			output, err := os.ReadFile(outFile.Name())
			if err != nil {
				return fmt.Errorf("failed to read converted file: %w", err)
			}

			result = bytes.NewReader(output)
			return nil

		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	})

	return result, err
}

// Create creates a new diagram programmatically.
func (r *D2Repository) Create(ctx context.Context, diagram *entity.Diagram) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Parse the content to create a graph.
	graph, _, err := d2compiler.Compile("", strings.NewReader(diagram.Content), &d2compiler.CompileOptions{
		UTF16Pos: false,
	})
	if err != nil {
		return fmt.Errorf("failed to compile diagram: %w", err)
	}

	r.diagrams[diagram.ID] = &diagramData{
		content: diagram.Content,
		graph:   graph,
	}

	return nil
}

// Export exports the diagram to the specified format.
func (r *D2Repository) Export(ctx context.Context, diagramID string, format entity.ExportFormat) (io.Reader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	// For now, use the stored content
	// TODO: Implement proper graph serialization to D2 text
	currentContent := data.content

	// Render the current state
	return r.Render(ctx, currentContent, format, nil)
}
