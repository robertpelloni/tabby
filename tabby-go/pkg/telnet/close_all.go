package telnet

// CloseAll closes all active Telnet connections.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	connIDs := make([]string, 0, len(m.connections))
	for id := range m.connections {
		connIDs = append(connIDs, id)
	}
	m.mu.Unlock()

	for _, id := range connIDs {
		_ = m.Close(id)
	}
}
