# D2 MCP Server

A Model Context Protocol (MCP) server that provides D2 diagram generation and manipulation capabilities.

D2 is a modern diagram scripting language that turns text to diagrams. This MCP server allows AI assistants like Claude to create, render, export, and save D2 diagrams programmatically.

The server provides 10 tools through the MCP protocol with enhanced descriptions for optimal AI assistant integration, enabling both simple diagram rendering and sophisticated incremental diagram building using the Oracle API.

With the new Oracle API integration, AI assistants can now build and modify diagrams incrementally, making it perfect for:
- Converting conversations into architecture diagrams
- Building flowcharts step-by-step as requirements are discussed
- Creating entity relationship diagrams from database schemas
- Generating system diagrams from code analysis
- Refining diagrams based on user feedback without starting over

## Features

### Basic Diagram Operations
- **d2_create** - Create new diagrams with optional initial content (unified approach)
- **d2_export** - Export diagrams to various formats (SVG, PNG, PDF)
- **d2_save** - Save existing diagrams to files

### Oracle API for Incremental Editing
- **d2_oracle_create** - Create shapes and connections incrementally
- **d2_oracle_set** - Set attributes on existing elements
- **d2_oracle_delete** - Delete specific elements from diagrams
- **d2_oracle_move** - Move shapes between containers
- **d2_oracle_rename** - Rename diagram elements
- **d2_oracle_get_info** - Get information about shapes, connections, or containers
- **d2_oracle_serialize** - Get the current D2 text representation of the diagram

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

### d2_create

Create a new diagram with optional initial content (unified approach):

**Empty diagram (for Oracle API workflow):**
```json
{
  "id": "my-diagram"
}
```

**With initial D2 content:**
```json
{
  "id": "my-diagram",
  "content": "a -> b: Hello\nserver: {shape: cylinder}"
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

#### d2_oracle_serialize

Get the current D2 text representation of the diagram:

```json
{
  "diagram_id": "my-diagram"
}
```

Returns the complete D2 text of the diagram including all modifications made through Oracle API.

### Creating Sequence Diagrams

D2 has built-in support for sequence diagrams. Use `d2_create` with proper D2 sequence diagram syntax:

```json
{
  "id": "api-flow",
  "content": "shape: sequence_diagram\n\nClient -> Server: HTTP Request\nServer -> Database: Query\nDatabase -> Server: Results\nServer -> Client: HTTP Response\n\n# Add styling\nClient -> Server.\"HTTP Request\": {style.stroke-dash: 3}\nDatabase -> Server.\"Results\": {style.stroke-dash: 3}"
}
```

**Example with actors and grouping:**
```json
{
  "id": "auth-flow",
  "content": "shape: sequence_diagram\n\ntitle: Authentication Flow {near: top-center}\n\n# Define actors\nClient: {shape: person}\nAuth Server: {shape: cloud}\nDatabase: {shape: cylinder}\n\n# Interactions\nClient -> Auth Server: Login Request\nAuth Server -> Database: Validate Credentials\nDatabase -> Auth Server: User Data\n\ngroup: Success Case {\n  Auth Server -> Client: Access Token\n  Client -> Auth Server: API Request + Token\n  Auth Server -> Client: API Response\n}\n\ngroup: Failure Case {\n  Auth Server -> Client: 401 Unauthorized\n}"
}
```

### Example Oracle API Workflow

**Starting from scratch:**
```javascript
// 1. Create an empty diagram
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

// 5. Export final result
d2_export({ diagramId: "architecture", format: "svg" })
```

**Starting with existing content (unified approach):**
```javascript
// 1. Create diagram with initial content
d2_create({ 
  id: "architecture",
  content: "web -> api -> db\ndb: {shape: cylinder}"
})

// 2. Enhance incrementally using Oracle API
d2_oracle_set({ diagram_id: "architecture", key: "web.label", value: "Web Server" })
d2_oracle_create({ diagram_id: "architecture", key: "cache" })
d2_oracle_create({ diagram_id: "architecture", key: "api -> cache" })

// 3. Export final result
d2_export({ diagramId: "architecture", format: "svg" })
```

### When to Use Each Tool

- **d2_create**: Always use for new diagrams - both empty (for incremental building) and with initial D2 content
- **d2_oracle_***: Use for incremental modifications to any diagram created with d2_create
- **d2_export**: Use to render the final diagram in your desired format

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

### v0.4.0 (Latest)
- Simplified API to unified `d2_create` for all diagram creation needs
- Enhanced tool descriptions for better AI assistant integration
- Improved Oracle API error handling and validation
- Reduced API surface from 14 to 10 tools
- **Breaking Change**: Removed d2_render, d2_render_to_file, d2_import, d2_from_text - use d2_create instead

### v0.3.0
- Added `d2_oracle_serialize` tool to get current D2 text representation

### v0.2.0
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
