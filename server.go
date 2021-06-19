package fws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type IEventHandler interface {
	React(msg *Msg)
}

type Server struct {
	mu      sync.RWMutex
	connMap map[string]*Conn
	handler IEventHandler
}

func NewServer(handler IEventHandler) *Server {
	return &Server{
		connMap: make(map[string]*Conn),
		handler: handler,
	}
}

func (s *Server) Serve(w http.ResponseWriter, r *http.Request, responseHeader http.Header) {
	u := websocket.Upgrader{
		// 允许所有CORS跨域请求
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	wsConn, err := u.Upgrade(w, r, responseHeader)
	if err != nil {
		panic(err)
	}
	conn := newConn(s, wsConn)
	s.addConn(conn)
	conn.start()
}

func (s *Server) addConn(conn *Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connMap[conn.connId] = conn
}

func (s *Server) removeConn(connId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connMap, connId)
}

func (s *Server) GetConn(connId string) *Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connMap[connId]
}

func (s *Server) GetAllConns() []*Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns := make([]*Conn, 0)
	for _, v := range s.connMap {
		conns = append(conns, v)
	}
	return conns
}

func (s *Server) GetConnsByMn(mn string) []*Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns := make([]*Conn, 0)
	for _, v := range s.connMap {
		if v.GetMn() == mn {
			conns = append(conns, v)
		}
	}
	return conns
}
