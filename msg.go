package fws

import "sync"

type Msg struct {
	mu   sync.RWMutex
	conn *Conn
	data []byte
}

func (m *Msg) GetConn() *Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn
}

func (m *Msg) GetData() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data
}
