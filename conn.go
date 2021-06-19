package fws

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
)

type Conn struct {
	mu      sync.RWMutex
	server  *Server
	wsConn  *websocket.Conn
	connId  string
	mn      string
	msgChan chan *Msg
	closed  bool
	ctx     context.Context
	cancel  context.CancelFunc
}

func newConn(server *Server, wsConn *websocket.Conn) *Conn {
	c := &Conn{
		server:  server,
		wsConn:  wsConn,
		connId:  uuid.New().String(),
		msgChan: make(chan *Msg),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

func (c *Conn) start() {
	go c.reader()
	go c.writer()
}

func (c *Conn) stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed == true {
		return
	}
	if err := c.wsConn.Close(); err != nil {
		return
	}
	c.closed = true
	c.cancel()
	c.server.removeConn(c.connId)
}

func (c *Conn) reader() {
	defer c.stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.receiveMsg()
			if err != nil {
				return
			}
			if msg == nil {
				continue
			}
			// 触发接收报文回调
			go c.server.handler.React(msg)
		}
	}
}

func (c *Conn) writer() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.msgChan:
			if err := c.wsConn.WriteMessage(1, msg.data); err != nil {
				return
			}
		}
	}
}

func (c *Conn) receiveMsg() (*Msg, error) {
	_, data, err := c.wsConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	msg := Msg{
		conn: c,
		data: data,
	}
	return &msg, nil
}

func (c *Conn) SendMsg(data []byte) error {
	c.mu.RLock()
	if c.closed == true {
		c.mu.RUnlock()
		return errors.New("connId " + c.connId + " conn closed when send msg")
	}
	c.mu.RUnlock()
	c.msgChan <- &Msg{
		data: data,
	}
	return nil
}

func (c *Conn) GetConnId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connId
}

func (c *Conn) GetMn() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mn
}

func (c *Conn) SetMn(mn string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.mn = mn
}

func (c *Conn) RemoteAddr() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.wsConn.RemoteAddr().String()
}
