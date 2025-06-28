package usecase

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/i2y/d2mcp/internal/domain/entity"
)

// Mock Oracle Repository for testing
type mockOracleRepository struct {
	// Control behavior
	shouldFail bool
	failMsg    string

	// Track calls
	createElementCalled bool
	setAttributeCalled  bool
	deleteElementCalled bool
	moveElementCalled   bool
	renameElementCalled bool
	getObjectCalled     bool
	getEdgeCalled       bool
	getChildrenCalled   bool
	loadDiagramCalled   bool
	serializeCalled     bool

	// Mock data
	mockObject   *entity.GraphObject
	mockEdge     *entity.GraphEdge
	mockChildren []string
}

func (m *mockOracleRepository) Render(ctx context.Context, content string, format entity.ExportFormat, theme *entity.Theme) (io.Reader, error) {
	return nil, nil
}

func (m *mockOracleRepository) Create(ctx context.Context, diagram *entity.Diagram) error {
	return nil
}

func (m *mockOracleRepository) Export(ctx context.Context, diagramID string, format entity.ExportFormat) (io.Reader, error) {
	return nil, nil
}

func (m *mockOracleRepository) CreateElement(ctx context.Context, diagramID string, boardPath []string, key string) (*entity.OracleResult, error) {
	m.createElementCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	return &entity.OracleResult{
		Success: true,
		NewKey:  key,
	}, nil
}

func (m *mockOracleRepository) SetAttribute(ctx context.Context, diagramID string, boardPath []string, key string, tag, value *string) (*entity.OracleResult, error) {
	m.setAttributeCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	return &entity.OracleResult{Success: true}, nil
}

func (m *mockOracleRepository) DeleteElement(ctx context.Context, diagramID string, boardPath []string, key string) (*entity.OracleResult, error) {
	m.deleteElementCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	return &entity.OracleResult{
		Success:  true,
		IDDeltas: map[string]string{"old": "new"},
	}, nil
}

func (m *mockOracleRepository) MoveElement(ctx context.Context, diagramID string, boardPath []string, key, newKey string, includeDescendants bool) (*entity.OracleResult, error) {
	m.moveElementCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	return &entity.OracleResult{Success: true}, nil
}

func (m *mockOracleRepository) RenameElement(ctx context.Context, diagramID string, boardPath []string, key, newName string) (*entity.OracleResult, error) {
	m.renameElementCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	return &entity.OracleResult{
		Success:  true,
		NewKey:   newName,
		IDDeltas: map[string]string{key: newName},
	}, nil
}

func (m *mockOracleRepository) GetObject(ctx context.Context, diagramID string, boardPath []string, objectID string) (*entity.GraphObject, error) {
	m.getObjectCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	if m.mockObject != nil {
		return m.mockObject, nil
	}
	return &entity.GraphObject{
		ID:    objectID,
		Label: "Test Object",
		Shape: "rectangle",
	}, nil
}

func (m *mockOracleRepository) GetEdge(ctx context.Context, diagramID string, boardPath []string, edgeID string) (*entity.GraphEdge, error) {
	m.getEdgeCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	if m.mockEdge != nil {
		return m.mockEdge, nil
	}
	return &entity.GraphEdge{
		ID:    edgeID,
		From:  "source",
		To:    "target",
		Label: "Test Edge",
	}, nil
}

func (m *mockOracleRepository) GetChildren(ctx context.Context, diagramID string, boardPath []string, parentID string) ([]string, error) {
	m.getChildrenCalled = true
	if m.shouldFail {
		return nil, errors.New(m.failMsg)
	}
	if m.mockChildren != nil {
		return m.mockChildren, nil
	}
	return []string{"child1", "child2"}, nil
}

func (m *mockOracleRepository) LoadDiagram(ctx context.Context, diagramID string, content string) error {
	m.loadDiagramCalled = true
	if m.shouldFail {
		return errors.New(m.failMsg)
	}
	return nil
}

func (m *mockOracleRepository) SerializeDiagram(ctx context.Context, diagramID string) (string, error) {
	m.serializeCalled = true
	if m.shouldFail {
		return "", errors.New(m.failMsg)
	}
	return "serialized content", nil
}

func TestOracleUseCase_CreateElement(t *testing.T) {
	tests := []struct {
		name       string
		op         *entity.OracleOperation
		shouldFail bool
		failMsg    string
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid create",
			op: &entity.OracleOperation{
				Type:      entity.OracleCreate,
				DiagramID: "test-diagram",
				Key:       "server",
			},
			shouldFail: false,
			wantErr:    false,
		},
		{
			name: "missing diagram ID",
			op: &entity.OracleOperation{
				Type: entity.OracleCreate,
				Key:  "server",
			},
			wantErr: true,
			errMsg:  "diagram ID is required",
		},
		{
			name: "missing key",
			op: &entity.OracleOperation{
				Type:      entity.OracleCreate,
				DiagramID: "test-diagram",
			},
			wantErr: true,
			errMsg:  "element key is required",
		},
		{
			name: "repository error",
			op: &entity.OracleOperation{
				Type:      entity.OracleCreate,
				DiagramID: "test-diagram",
				Key:       "server",
			},
			shouldFail: true,
			failMsg:    "repository error",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOracleRepository{
				shouldFail: tt.shouldFail,
				failMsg:    tt.failMsg,
			}
			uc := NewOracleUseCase(mockRepo)

			result, err := uc.CreateElement(context.Background(), tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("CreateElement() error = %v, want %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr && !mockRepo.createElementCalled {
				t.Error("CreateElement() repository method not called")
			}
			if !tt.wantErr && result != nil && !result.Success {
				t.Error("CreateElement() Success = false")
			}
		})
	}
}

func TestOracleUseCase_SetAttribute(t *testing.T) {
	value := "cylinder"

	tests := []struct {
		name    string
		op      *entity.OracleOperation
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid set attribute",
			op: &entity.OracleOperation{
				Type:      entity.OracleSet,
				DiagramID: "test-diagram",
				Key:       "server.shape",
				Value:     &value,
			},
			wantErr: false,
		},
		{
			name: "missing diagram ID",
			op: &entity.OracleOperation{
				Type:  entity.OracleSet,
				Key:   "server.shape",
				Value: &value,
			},
			wantErr: true,
			errMsg:  "diagram ID is required",
		},
		{
			name: "missing key",
			op: &entity.OracleOperation{
				Type:      entity.OracleSet,
				DiagramID: "test-diagram",
				Value:     &value,
			},
			wantErr: true,
			errMsg:  "element key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOracleRepository{}
			uc := NewOracleUseCase(mockRepo)

			result, err := uc.SetAttribute(context.Background(), tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAttribute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("SetAttribute() error = %v, want %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr && !mockRepo.setAttributeCalled {
				t.Error("SetAttribute() repository method not called")
			}
			if !tt.wantErr && result != nil && !result.Success {
				t.Error("SetAttribute() Success = false")
			}
		})
	}
}

func TestOracleUseCase_MoveElement(t *testing.T) {
	newKey := "container.server"

	tests := []struct {
		name    string
		op      *entity.OracleOperation
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid move",
			op: &entity.OracleOperation{
				Type:               entity.OracleMove,
				DiagramID:          "test-diagram",
				Key:                "server",
				NewKey:             &newKey,
				IncludeDescendants: true,
			},
			wantErr: false,
		},
		{
			name: "missing new key",
			op: &entity.OracleOperation{
				Type:      entity.OracleMove,
				DiagramID: "test-diagram",
				Key:       "server",
			},
			wantErr: true,
			errMsg:  "new key is required",
		},
		{
			name: "empty new key",
			op: &entity.OracleOperation{
				Type:      entity.OracleMove,
				DiagramID: "test-diagram",
				Key:       "server",
				NewKey:    stringPtr(""),
			},
			wantErr: true,
			errMsg:  "new key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOracleRepository{}
			uc := NewOracleUseCase(mockRepo)

			result, err := uc.MoveElement(context.Background(), tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoveElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("MoveElement() error = %v, want %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr && !mockRepo.moveElementCalled {
				t.Error("MoveElement() repository method not called")
			}
			if !tt.wantErr && result != nil && !result.Success {
				t.Error("MoveElement() Success = false")
			}
		})
	}
}

func TestOracleUseCase_ExecuteOperation(t *testing.T) {
	value := "test"
	newKey := "newname"

	tests := []struct {
		name    string
		op      *entity.OracleOperation
		wantErr bool
	}{
		{
			name: "execute create",
			op: &entity.OracleOperation{
				Type:      entity.OracleCreate,
				DiagramID: "test",
				Key:       "server",
			},
			wantErr: false,
		},
		{
			name: "execute set",
			op: &entity.OracleOperation{
				Type:      entity.OracleSet,
				DiagramID: "test",
				Key:       "server.label",
				Value:     &value,
			},
			wantErr: false,
		},
		{
			name: "execute delete",
			op: &entity.OracleOperation{
				Type:      entity.OracleDelete,
				DiagramID: "test",
				Key:       "server",
			},
			wantErr: false,
		},
		{
			name: "execute move",
			op: &entity.OracleOperation{
				Type:      entity.OracleMove,
				DiagramID: "test",
				Key:       "server",
				NewKey:    &newKey,
			},
			wantErr: false,
		},
		{
			name: "execute rename",
			op: &entity.OracleOperation{
				Type:      entity.OracleRename,
				DiagramID: "test",
				Key:       "server",
				NewKey:    &newKey,
			},
			wantErr: false,
		},
		{
			name: "unknown operation type",
			op: &entity.OracleOperation{
				Type:      "unknown",
				DiagramID: "test",
				Key:       "server",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOracleRepository{}
			uc := NewOracleUseCase(mockRepo)

			_, err := uc.ExecuteOperation(context.Background(), tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOracleUseCase_LoadAndSerialize(t *testing.T) {
	tests := []struct {
		name      string
		diagramID string
		content   string
		wantErr   bool
	}{
		{
			name:      "valid load",
			diagramID: "test",
			content:   "a -> b",
			wantErr:   false,
		},
		{
			name:      "empty diagram ID",
			diagramID: "",
			content:   "a -> b",
			wantErr:   true,
		},
		{
			name:      "empty content",
			diagramID: "test",
			content:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOracleRepository{}
			uc := NewOracleUseCase(mockRepo)

			err := uc.LoadDiagram(context.Background(), tt.diagramID, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadDiagram() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !mockRepo.loadDiagramCalled {
				t.Error("LoadDiagram() repository method not called")
			}
		})
	}

	// Test SerializeDiagram
	t.Run("serialize diagram", func(t *testing.T) {
		mockRepo := &mockOracleRepository{}
		uc := NewOracleUseCase(mockRepo)

		content, err := uc.SerializeDiagram(context.Background(), "test")
		if err != nil {
			t.Errorf("SerializeDiagram() error = %v", err)
		}
		if !mockRepo.serializeCalled {
			t.Error("SerializeDiagram() repository method not called")
		}
		if content == "" {
			t.Error("SerializeDiagram() returned empty content")
		}
	})

	t.Run("serialize with empty ID", func(t *testing.T) {
		mockRepo := &mockOracleRepository{}
		uc := NewOracleUseCase(mockRepo)

		_, err := uc.SerializeDiagram(context.Background(), "")
		if err == nil {
			t.Error("SerializeDiagram() should fail with empty ID")
		}
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
