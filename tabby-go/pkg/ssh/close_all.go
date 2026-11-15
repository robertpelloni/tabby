package ssh

import (
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// CloseAll closes all active SSH connections, sessions, and port forwards.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	connIDs := make([]string, 0, len(m.connections))
	for id := range m.connections {
		connIDs = append(connIDs, id)
	}
	m.mu.Unlock()

	for _, id := range connIDs {
		_ = m.Close(api.SSHCloseParams{ConnectionID: id})
	}
}
