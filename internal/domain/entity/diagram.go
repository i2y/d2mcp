package entity

// Diagram represents a D2 diagram entity.
type Diagram struct {
	ID      string
	Content string
	Format  ExportFormat
	Theme   *Theme
}

// ExportFormat represents the output format for diagram export.
type ExportFormat string

const (
	// FormatSVG represents SVG export format.
	FormatSVG ExportFormat = "svg"
	// FormatPNG represents PNG export format.
	FormatPNG ExportFormat = "png"
	// FormatPDF represents PDF export format.
	FormatPDF ExportFormat = "pdf"
)

// Theme represents a D2 diagram theme.
type Theme struct {
	ID   int
	Name string
}

// Shape represents a shape in a D2 diagram.
type Shape struct {
	ID         string
	Label      string
	Attributes map[string]string
}

// Connection represents a connection between shapes.
type Connection struct {
	From       string
	To         string
	Label      string
	Attributes map[string]string
}
