package entity

// OracleOperation represents a diagram manipulation operation
type OracleOperation struct {
	Type               OracleOperationType
	DiagramID          string
	BoardPath          []string // For multi-board support
	Key                string
	Value              *string
	Tag                *string
	NewKey             *string
	IncludeDescendants bool
}

// OracleOperationType defines the type of oracle operation
type OracleOperationType string

const (
	OracleCreate OracleOperationType = "create"
	OracleSet    OracleOperationType = "set"
	OracleDelete OracleOperationType = "delete"
	OracleMove   OracleOperationType = "move"
	OracleRename OracleOperationType = "rename"
)

// OracleResult represents the result of an oracle operation
type OracleResult struct {
	Success  bool
	NewKey   string
	IDDeltas map[string]string // Maps old IDs to new IDs
	Graph    *DiagramGraph     // The resulting graph state
}

// DiagramGraph represents the internal graph structure
type DiagramGraph struct {
	ID      string
	Content string
	Objects map[string]*GraphObject
	Edges   map[string]*GraphEdge
}

// GraphObject represents a shape in the diagram
type GraphObject struct {
	ID         string
	Label      string
	Shape      string
	Parent     string
	Attributes map[string]interface{}
}

// GraphEdge represents a connection in the diagram
type GraphEdge struct {
	ID         string
	From       string
	To         string
	Label      string
	Attributes map[string]interface{}
}
