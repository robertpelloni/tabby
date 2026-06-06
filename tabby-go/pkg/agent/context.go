package agent

import (
	"sync"
	"time"
)

// FileContext represents a file and its relevant metadata for the agent
type FileContext struct {
	Path        string    `json:"path"`
	LastRead    time.Time `json:"lastRead"`
	Summary     string    `json:"summary,omitempty"`
	IsImportant bool      `json:"isImportant"`
}

// ContextManager manages the agent's long-term and short-term memory
type ContextManager struct {
	files   map[string]*FileContext
	symbols map[string]string // Maps symbol names to file paths
	mu      sync.RWMutex
}

func NewContextManager() *ContextManager {
	return &ContextManager{
		files:   make(map[string]*FileContext),
		symbols: make(map[string]string),
	}
}

func (cm *ContextManager) MarkFileRead(path string, summary string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.files[path] = &FileContext{
		Path:     path,
		LastRead: time.Now(),
		Summary:  summary,
	}
}

func (cm *ContextManager) RegisterSymbol(name string, path string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.symbols[name] = path
}

func (cm *ContextManager) GetFileContext(path string) (*FileContext, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	ctx, ok := cm.files[path]
	return ctx, ok
}

func (cm *ContextManager) GetFiles() []*FileContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	files := make([]*FileContext, 0, len(cm.files))
	for _, f := range cm.files {
		files = append(files, f)
	}
	return files
}
