# D2 MCP Server

A Model Context Protocol (MCP) server that provides D2 diagram generation and manipulation capabilities.

D2 is a modern diagram scripting language that turns text to diagrams. This MCP server allows AI assistants like Claude to create, render, export, and save D2 diagrams programmatically.

With the new Oracle API integration, AI assistants can now build and modify diagrams incrementally, making it perfect for:
- Converting conversations into architecture diagrams
- Building flowcharts step-by-step as requirements are discussed
- Creating entity relationship diagrams from database schemas
- Generating system diagrams from code analysis
- Refining diagrams based on user feedback without starting over

## Features

### Basic Diagram Operations
- **d2_render** - Render D2 text into diagrams (SVG, PNG, PDF formats)
- **d2_render_to_file** - Render D2 text and save directly to file
- **d2_create** - Create new diagrams programmatically  
- **d2_export** - Export diagrams to various formats (SVG, PNG, PDF)
- **d2_save** - Save existing diagrams to files

### Oracle API for Incremental Editing (New!)
- **d2_oracle_create** - Create shapes and connections incrementally
- **d2_oracle_set** - Set attributes on existing elements
- **d2_oracle_delete** - Delete specific elements from diagrams
- **d2_oracle_move** - Move shapes between containers
- **d2_oracle_rename** - Rename diagram elements
- **d2_oracle_get_info** - Get information about shapes, connections, or containers

### Additional Features
- **20 themes** - Support for all D2 themes (18 light + 2 dark)
- **MCP Protocol** - Standard protocol for AI tool integration

## Project Structure

```
d2mcp/
├── cmd/                  # Application entry point
├── internal/
│   ├── domain/          # Business entities and interfaces
│   │   ├── entity/      # Domain entities
│   │   └── repository/  # Repository interfaces
│   ├── usecase/         # Business logic
│   ├── infrastructure/  # External implementations
│   │   ├── d2/          # D2 library integration
│   │   └── mcp/         # MCP server implementation
│   └── presentation/    # MCP handlers
│       └── handler/     # Tool handlers
└── pkg/                 # Public packages
```

## Prerequisites

- Go 1.24.3 or higher
- D2 v0.6.7 or higher (included as dependency)
- For PNG/PDF export (optional):
  - `rsvg-convert` (from librsvg) or
  - ImageMagick (`convert` command)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/i2y/d2mcp.git
cd d2mcp

# Build the binary
make build

# Or build for all platforms
make build-all
```

### Using Go Install

```bash
go install github.com/i2y/d2mcp/cmd@latest
```

## Building

```bash
# Simple build
make build

# Run directly
make run

# Cross-platform builds
make build-all
```

## Usage

### With Claude Desktop

Add to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "d2mcp": {
      "command": "/path/to/d2mcp"
    }
  }
}
```

Replace `/path/to/d2mcp` with the actual path to your built binary.

### Standalone

```bash
# Run the MCP server (stdio transport)
./d2mcp
```

## Tools

### d2_render

Render D2 text into a diagram:

```json
{
  "content": "a -> b: Hello",
  "format": "svg",  // Options: "svg", "png", "pdf" (default: "svg")
  "theme": 0        // Theme ID (see THEMES.md for full list, default: 0)
}
```

**Example D2 content**:
```
# Basic flow
start -> middle: Request
middle -> end: Response

# With styling
server: {
  shape: cylinder
  style.fill: "#f0f0f0"
}
```

**Note**: PNG and PDF formats require either `rsvg-convert` or ImageMagick (`convert`) to be installed on your system.

### d2_render_to_file

Render D2 text and save to a file:

```json
{
  "content": "a -> b: Hello",
  "format": "png",
  "theme": 0
}
```

Returns the path to the saved file.

### d2_create

Create a new diagram:

```json
{
  "id": "my-diagram",
  "content": "initial content (optional)"
}
```

### d2_export

Export a diagram to a specific format:

```json
{
  "diagramId": "my-diagram",
  "format": "png"  // Options: "svg", "png", "pdf"
}
```

### d2_save

Save a diagram to a file:

```json
{
  "diagramId": "my-diagram",
  "format": "pdf",
  "path": "/path/to/output.pdf"  // Optional, defaults to temp directory
}
```

### Oracle API Tools

The Oracle API tools enable incremental diagram manipulation without regenerating the entire diagram. These tools are ideal for building diagrams step-by-step or making surgical edits.

#### d2_oracle_create

Create a new shape or connection:

```json
{
  "diagram_id": "my-diagram",
  "key": "server"  // Creates a shape
}
```

```json
{
  "diagram_id": "my-diagram", 
  "key": "server -> database"  // Creates a connection
}
```

#### d2_oracle_set

Set attributes on existing elements:

```json
{
  "diagram_id": "my-diagram",
  "key": "server.shape",
  "value": "cylinder"
}
```

```json
{
  "diagram_id": "my-diagram",
  "key": "server.style.fill",
  "value": "#f0f0f0"
}
```

#### d2_oracle_delete

Delete elements from the diagram:

```json
{
  "diagram_id": "my-diagram",
  "key": "server"  // Deletes the server and its children
}
```

#### d2_oracle_move

Move elements between containers:

```json
{
  "diagram_id": "my-diagram",
  "key": "server",
  "new_parent": "network.internal",  // Moves server into network.internal
  "include_descendants": "true"       // Also moves child elements
}
```

#### d2_oracle_rename

Rename diagram elements:

```json
{
  "diagram_id": "my-diagram",
  "key": "server",
  "new_name": "web_server"
}
```

#### d2_oracle_get_info

Get information about diagram elements:

```json
{
  "diagram_id": "my-diagram",
  "key": "server",
  "info_type": "object"  // Options: "object", "edge", "children"
}
```

### Example Oracle API Workflow

```javascript
// 1. Create a diagram
d2_create({ id: "architecture" })

// 2. Add shapes incrementally
d2_oracle_create({ diagram_id: "architecture", key: "web" })
d2_oracle_create({ diagram_id: "architecture", key: "api" })
d2_oracle_create({ diagram_id: "architecture", key: "db" })

// 3. Set properties
d2_oracle_set({ diagram_id: "architecture", key: "db.shape", value: "cylinder" })
d2_oracle_set({ diagram_id: "architecture", key: "web.label", value: "Web Server" })

// 4. Create connections
d2_oracle_create({ diagram_id: "architecture", key: "web -> api" })
d2_oracle_create({ diagram_id: "architecture", key: "api -> db" })

// 5. Reorganize if needed
d2_oracle_move({ diagram_id: "architecture", key: "api", new_parent: "backend" })

// 6. Export final result
d2_export({ diagramId: "architecture", format: "svg" })
```

## Development

### Running tests

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/presentation/handler
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

### Adding new features

1. Define entities in `internal/domain/entity`
2. Add repository interfaces in `internal/domain/repository`
3. Implement business logic in `internal/usecase`
4. Add infrastructure implementations
5. Create MCP handlers in `internal/presentation/handler`
6. Wire dependencies in `cmd/main.go`

### Project Structure

- **cmd/**: Application entry point
- **internal/domain/**: Core business logic and entities
- **internal/infrastructure/**: External service integrations
- **internal/presentation/**: MCP protocol handlers
- **internal/usecase/**: Application business logic

## Troubleshooting

### PNG/PDF Export Not Working

If you get errors when exporting to PNG or PDF formats, install one of these tools:

**macOS**:
```bash
# Using Homebrew
brew install librsvg
# or
brew install imagemagick
```

**Ubuntu/Debian**:
```bash
sudo apt-get install librsvg2-bin
# or
sudo apt-get install imagemagick
```

**Windows**:
Download and install ImageMagick from the official website.

### MCP Connection Issues

1. Ensure the binary has execute permissions: `chmod +x d2mcp`
2. Check Claude Desktop logs for error messages
3. Verify the path in your configuration is absolute

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Changelog

### v0.2.0 (Latest)
- Added D2 Oracle API integration for incremental diagram manipulation
- 6 new MCP tools for creating, modifying, and querying diagram elements
- Support for stateful diagram editing sessions

### v0.1.0
- Initial release with basic D2 diagram operations
- Support for rendering, creating, exporting, and saving diagrams
- 20 built-in themes
- MCP protocol integration

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
