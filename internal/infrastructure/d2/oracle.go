package d2

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2oracle"

	"github.com/i2y/d2mcp/internal/domain/entity"
	"github.com/i2y/d2mcp/internal/domain/repository"
)

// OracleSession represents an active Oracle editing session
type OracleSession struct {
	DiagramID    string
	Graph        *d2graph.Graph
	AST          *d2ast.Map
	LastModified time.Time
	Operations   []entity.OracleOperation // History tracking
}

// D2OracleRepository extends D2Repository with Oracle capabilities
type D2OracleRepository struct {
	*D2Repository
	sessions  map[string]*OracleSession
	sessionMu sync.RWMutex
}

// NewD2OracleRepository creates a new D2 repository with Oracle support
func NewD2OracleRepository() repository.OracleRepository {
	return &D2OracleRepository{
		D2Repository: &D2Repository{
			diagrams: make(map[string]*diagramData),
		},
		sessions: make(map[string]*OracleSession),
	}
}

// LoadDiagram loads a diagram from D2 text
func (r *D2OracleRepository) LoadDiagram(ctx context.Context, diagramID string, content string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Parse the content to create a graph
	graph, _, err := d2compiler.Compile("", strings.NewReader(content), &d2compiler.CompileOptions{
		UTF16Pos: false,
	})
	if err != nil {
		return fmt.Errorf("failed to compile diagram: %w", err)
	}

	r.diagrams[diagramID] = &diagramData{
		content: content,
		graph:   graph,
	}

	return nil
}

// SerializeDiagram converts the current graph state back to D2 text
func (r *D2OracleRepository) SerializeDiagram(ctx context.Context, diagramID string) (string, error) {
	r.mu.RLock()
	data, exists := r.diagrams[diagramID]
	r.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("diagram %s not found", diagramID)
	}

	// Check if there's an active session with a modified graph
	r.sessionMu.RLock()
	session, hasSession := r.sessions[diagramID]
	r.sessionMu.RUnlock()

	var graph *d2graph.Graph
	if hasSession && session.Graph != nil {
		graph = session.Graph
	} else {
		graph = data.graph
	}

	// Convert graph back to D2 text
	if graph.AST == nil {
		return data.content, nil // Return original content if no AST
	}

	// Format the AST back to D2 text
	formatted := d2format.Format(graph.AST)
	return formatted, nil
}

// CreateElement creates a new shape or connection
func (r *D2OracleRepository) CreateElement(ctx context.Context, diagramID string, boardPath []string, key string) (*entity.OracleResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	// Get or create session
	session := r.getOrCreateSession(diagramID, data.graph)

	// Use d2oracle to create element
	newGraph, newKey, err := d2oracle.Create(session.Graph, boardPath, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create element: %w", err)
	}

	// Update session
	session.Graph = newGraph
	session.LastModified = time.Now()

	// Update stored graph
	data.graph = newGraph

	// Serialize back to D2 text
	if newGraph.AST != nil {
		data.content = d2format.Format(newGraph.AST)
	}

	return &entity.OracleResult{
		Success: true,
		NewKey:  newKey,
		Graph:   r.graphToEntity(newGraph),
	}, nil
}

// SetAttribute sets attributes on a shape or connection
func (r *D2OracleRepository) SetAttribute(ctx context.Context, diagramID string, boardPath []string, key string, tag, value *string) (*entity.OracleResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	session := r.getOrCreateSession(diagramID, data.graph)

	// Use d2oracle to set attribute
	newGraph, err := d2oracle.Set(session.Graph, boardPath, key, tag, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set attribute: %w", err)
	}

	// Update session and stored graph
	session.Graph = newGraph
	session.LastModified = time.Now()
	data.graph = newGraph

	// Serialize back to D2 text
	if newGraph.AST != nil {
		data.content = d2format.Format(newGraph.AST)
	}

	return &entity.OracleResult{
		Success: true,
		Graph:   r.graphToEntity(newGraph),
	}, nil
}

// DeleteElement deletes a shape or connection
func (r *D2OracleRepository) DeleteElement(ctx context.Context, diagramID string, boardPath []string, key string) (*entity.OracleResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	session := r.getOrCreateSession(diagramID, data.graph)

	// Check if this is a connection deletion (contains "->")
	isConnection := strings.Contains(key, "->")

	// Try to get ID deltas before deletion, but handle panic gracefully
	var idDeltas map[string]string
	func() {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				// If panic occurs, just use empty ID deltas
				idDeltas = make(map[string]string)
			}
		}()
		var err error
		idDeltas, err = d2oracle.DeleteIDDeltas(session.Graph, boardPath, key)
		if err != nil {
			idDeltas = make(map[string]string)
		}
	}()

	// Use d2oracle to delete element
	var newGraph *d2graph.Graph
	var deleteErr error
	func() {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				// If panic occurs during delete, try alternative approach for connections
				if isConnection {
					// For connections, we'll recreate the graph without this connection
					deleteErr = r.deleteConnectionWorkaround(ctx, diagramID, key)
					if deleteErr == nil {
						// Reload the graph after workaround
						newGraph = r.diagrams[diagramID].graph
					}
				} else {
					deleteErr = fmt.Errorf("failed to delete element: panic occurred - %v", panicErr)
				}
			}
		}()
		if deleteErr == nil {
			newGraph, deleteErr = d2oracle.Delete(session.Graph, boardPath, key)
		}
	}()

	if deleteErr != nil {
		return nil, deleteErr
	}

	// Update session and stored graph
	session.Graph = newGraph
	session.LastModified = time.Now()
	data.graph = newGraph

	// Serialize back to D2 text
	if newGraph.AST != nil {
		data.content = d2format.Format(newGraph.AST)
	}

	return &entity.OracleResult{
		Success:  true,
		IDDeltas: idDeltas,
		Graph:    r.graphToEntity(newGraph),
	}, nil
}

// MoveElement moves a shape to a new container
func (r *D2OracleRepository) MoveElement(ctx context.Context, diagramID string, boardPath []string, key, newKey string, includeDescendants bool) (*entity.OracleResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	session := r.getOrCreateSession(diagramID, data.graph)

	// Use d2oracle to move element
	newGraph, err := d2oracle.Move(session.Graph, boardPath, key, newKey, includeDescendants)
	if err != nil {
		return nil, fmt.Errorf("failed to move element: %w", err)
	}

	// Update session and stored graph
	session.Graph = newGraph
	session.LastModified = time.Now()
	data.graph = newGraph

	// Serialize back to D2 text
	if newGraph.AST != nil {
		data.content = d2format.Format(newGraph.AST)
	}

	return &entity.OracleResult{
		Success: true,
		Graph:   r.graphToEntity(newGraph),
	}, nil
}

// RenameElement renames a shape or connection
func (r *D2OracleRepository) RenameElement(ctx context.Context, diagramID string, boardPath []string, key, newName string) (*entity.OracleResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	session := r.getOrCreateSession(diagramID, data.graph)

	// Get ID deltas before rename
	idDeltas, err := d2oracle.RenameIDDeltas(session.Graph, boardPath, key, newName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rename ID deltas: %w", err)
	}

	// Use d2oracle to rename element
	newGraph, newRenamedKey, err := d2oracle.Rename(session.Graph, boardPath, key, newName)
	if err != nil {
		return nil, fmt.Errorf("failed to rename element: %w", err)
	}

	// Update session and stored graph
	session.Graph = newGraph
	session.LastModified = time.Now()
	data.graph = newGraph

	// Serialize back to D2 text
	if newGraph.AST != nil {
		data.content = d2format.Format(newGraph.AST)
	}

	return &entity.OracleResult{
		Success:  true,
		NewKey:   newRenamedKey,
		IDDeltas: idDeltas,
		Graph:    r.graphToEntity(newGraph),
	}, nil
}

// GetObject retrieves object information
func (r *D2OracleRepository) GetObject(ctx context.Context, diagramID string, boardPath []string, objectID string) (*entity.GraphObject, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	// Use d2oracle to get object
	obj := d2oracle.GetObj(data.graph, boardPath, objectID)
	if obj == nil {
		return nil, fmt.Errorf("object %s not found", objectID)
	}

	return r.objectToEntity(obj), nil
}

// GetEdge retrieves edge information
func (r *D2OracleRepository) GetEdge(ctx context.Context, diagramID string, boardPath []string, edgeID string) (*entity.GraphEdge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	// Use d2oracle to get edge
	edge := d2oracle.GetEdge(data.graph, boardPath, edgeID)
	if edge == nil {
		return nil, fmt.Errorf("edge %s not found", edgeID)
	}

	return r.edgeToEntity(edge), nil
}

// GetChildren retrieves child element IDs
func (r *D2OracleRepository) GetChildren(ctx context.Context, diagramID string, boardPath []string, parentID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.diagrams[diagramID]
	if !exists {
		return nil, fmt.Errorf("diagram %s not found", diagramID)
	}

	// Use d2oracle to get children IDs
	childrenIDs, err := d2oracle.GetChildrenIDs(data.graph, boardPath, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children IDs: %w", err)
	}
	return childrenIDs, nil
}

// Helper methods

func (r *D2OracleRepository) getOrCreateSession(diagramID string, graph *d2graph.Graph) *OracleSession {
	r.sessionMu.Lock()
	defer r.sessionMu.Unlock()

	if session, exists := r.sessions[diagramID]; exists {
		return session
	}

	session := &OracleSession{
		DiagramID:    diagramID,
		Graph:        graph,
		LastModified: time.Now(),
		Operations:   []entity.OracleOperation{},
	}

	r.sessions[diagramID] = session
	return session
}

// deleteConnectionWorkaround handles connection deletion when Oracle API panics
// Note: This method assumes the caller already holds the mutex lock
func (r *D2OracleRepository) deleteConnectionWorkaround(ctx context.Context, diagramID string, connectionKey string) error {
	// Get current data - no lock needed as caller already has it
	data, exists := r.diagrams[diagramID]
	if !exists {
		return fmt.Errorf("diagram %s not found", diagramID)
	}

	// Get current D2 text from the data
	currentD2 := data.content
	if data.graph != nil && data.graph.AST != nil {
		currentD2 = d2format.Format(data.graph.AST)
	}

	// Parse the connection key (e.g., "Customer -> Order")
	parts := strings.Split(connectionKey, "->")
	if len(parts) != 2 {
		return fmt.Errorf("invalid connection key format: %s", connectionKey)
	}

	src := strings.TrimSpace(parts[0])
	dst := strings.TrimSpace(parts[1])

	// Remove the connection from D2 text
	// This is a simple approach - for production, we'd want to use proper AST manipulation
	lines := strings.Split(currentD2, "\n")
	var newLines []string
	for _, line := range lines {
		// Skip lines that match the connection pattern
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, src+" -> "+dst) ||
			strings.HasPrefix(trimmedLine, src+"-> "+dst) ||
			strings.HasPrefix(trimmedLine, src+" ->"+dst) ||
			strings.HasPrefix(trimmedLine, src+"->"+dst) {
			continue
		}
		newLines = append(newLines, line)
	}

	newD2 := strings.Join(newLines, "\n")

	// Compile the new D2 text directly without calling LoadDiagram to avoid deadlock
	graph, _, err := d2compiler.Compile("", strings.NewReader(newD2), &d2compiler.CompileOptions{
		UTF16Pos: false,
	})
	if err != nil {
		return fmt.Errorf("failed to compile diagram after removing connection: %w", err)
	}

	// Update the diagram data directly
	r.diagrams[diagramID] = &diagramData{
		content: newD2,
		graph:   graph,
	}

	return nil
}

func (r *D2OracleRepository) graphToEntity(graph *d2graph.Graph) *entity.DiagramGraph {
	if graph == nil {
		return nil
	}

	diagramGraph := &entity.DiagramGraph{
		Objects: make(map[string]*entity.GraphObject),
		Edges:   make(map[string]*entity.GraphEdge),
	}

	// Convert objects
	for _, obj := range graph.Objects {
		diagramGraph.Objects[obj.ID] = r.objectToEntity(obj)
	}

	// Convert edges
	for _, edge := range graph.Edges {
		edgeEntity := r.edgeToEntity(edge)
		if edgeEntity != nil {
			diagramGraph.Edges[edgeEntity.ID] = edgeEntity
		}
	}

	return diagramGraph
}

func (r *D2OracleRepository) objectToEntity(obj *d2graph.Object) *entity.GraphObject {
	if obj == nil {
		return nil
	}

	graphObj := &entity.GraphObject{
		ID:         obj.ID,
		Label:      obj.Label.Value,
		Attributes: make(map[string]interface{}),
	}

	if obj.Shape.Value != "" {
		graphObj.Shape = obj.Shape.Value
	}

	if obj.Parent != nil {
		graphObj.Parent = obj.Parent.ID
	}

	// Convert key attributes
	if obj.Attributes.Label.Value != "" {
		graphObj.Attributes["label"] = obj.Attributes.Label.Value
	}
	if obj.Attributes.Shape.Value != "" {
		graphObj.Attributes["shape"] = obj.Attributes.Shape.Value
	}
	if obj.Attributes.Style.Fill != nil && obj.Attributes.Style.Fill.Value != "" {
		graphObj.Attributes["fill"] = obj.Attributes.Style.Fill.Value
	}
	if obj.Attributes.Style.Stroke != nil && obj.Attributes.Style.Stroke.Value != "" {
		graphObj.Attributes["stroke"] = obj.Attributes.Style.Stroke.Value
	}

	return graphObj
}

func (r *D2OracleRepository) edgeToEntity(edge *d2graph.Edge) *entity.GraphEdge {
	if edge == nil {
		return nil
	}

	// Generate ID from source and destination
	edgeID := fmt.Sprintf("%d", edge.Index)
	if edge.Src != nil && edge.Dst != nil {
		edgeID = fmt.Sprintf("%s->%s", edge.Src.ID, edge.Dst.ID)
	}

	graphEdge := &entity.GraphEdge{
		ID:         edgeID,
		Label:      edge.Label.Value,
		Attributes: make(map[string]interface{}),
	}

	if edge.Src != nil {
		graphEdge.From = edge.Src.ID
	}

	if edge.Dst != nil {
		graphEdge.To = edge.Dst.ID
	}

	// Convert key attributes
	if edge.Attributes.Label.Value != "" {
		graphEdge.Attributes["label"] = edge.Attributes.Label.Value
	}
	if edge.Attributes.Style.Stroke != nil && edge.Attributes.Style.Stroke.Value != "" {
		graphEdge.Attributes["stroke"] = edge.Attributes.Style.Stroke.Value
	}
	if edge.SrcArrow {
		graphEdge.Attributes["srcArrow"] = true
	}
	if edge.DstArrow {
		graphEdge.Attributes["dstArrow"] = true
	}

	return graphEdge
}
