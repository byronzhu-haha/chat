package conn

import (
	"context"
	"errors"
	"github.com/byronzhu-haha/chat/entity/message"
	"github.com/byronzhu-haha/chat/server/repo"
	"github.com/byronzhu-haha/log"
	"net"
	"sync"
	"time"
)

const sendAndRecvG = 1000

type Manager struct {
	init    bool
	conns   map[string]*Conn
	mu      sync.RWMutex
	stop    chan struct{}
	postman chan []byte
	metaCh  chan message.Message
}

type Conn struct {
	conn   net.Conn
	reader chan []byte
	stop   chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		init:    true,
		conns:   make(map[string]*Conn),
		stop:    make(chan struct{}),
		postman: make(chan []byte, sendAndRecvG),
		metaCh:  make(chan message.Message, sendAndRecvG),
	}
}

func (m *Manager) Start() error {
	if !m.init {
		return errors.New("manager of conn is not init")
	}
	listener, err := net.Listen("tcp", ":4567")
	if err != nil {
		return err
	}
	go m.accept(listener)
	return nil
}

func (m *Manager) accept(listener net.Listener) {
	for {
		if m.check() {
			break
		}

		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("accept failed, err: %+v", err)
			continue
		}
		c := newConn(conn)
		m.mu.Lock()
		m.conns[conn.RemoteAddr().String()] = c
		m.mu.Unlock()

		c.start(m.postman)
	}
}

func (m *Manager) transferMsg() {
	for bytes := range m.postman {
		if m.check() {
			return
		}
		msg, err := message.Unpack(bytes)
		if err != nil {
			log.Errorf("unpack message failed, err: %+v", err)
			continue
		}
		if msg.IsRequestMsg() {
			m.sendMetadata(msg)
			continue
		}
		if !msg.IsChatMsg() {
			log.Warnf("invalid msg type, it should be chat msg")
			continue
		}
		head, err := message.UnpackChatHeader(msg.Head)
		if err != nil {
			log.Errorf("unmarshal header failed, err: %+v", err)
			continue
		}
		addr, err := repo.GetUserIP(head.DestUserID)
		if err != nil {
			log.Errorf("get ip for user(%s) failed, err: %+v", head.DestUserID, err)
		}
		m.mu.RLock()
		conn, ok := m.conns[addr]
		m.mu.RUnlock()
		if !ok {
			log.Infof("user is not online, receive: %s, sender: %d", head.DestUserID, head.SrcUserID)
			continue
		}

		conn.write(bytes)
	}
}

func (m *Manager) Broadcast(msg []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, conn := range m.conns {
		conn.write(msg)
	}
}

func (m *Manager) SendMsg(addr string, data []byte) {
	m.mu.RLock()
	conn, ok := m.conns[addr]
	m.mu.RUnlock()
	if !ok {
		return
	}
	conn.write(data)
}

func (m *Manager) HandleMetadata(ctx context.Context, receiver chan<- message.Message) {
	go func() {
		for meta := range m.metaCh {
			if m.check() {
				break
			}
			select {
			case <-ctx.Done():
				return
			case receiver <- meta:
			}
		}
	}()
}

func (m *Manager) sendMetadata(msg message.Message) {
	go func() {
		if m.check() {
			return
		}
		m.metaCh <- msg
	}()
}

func (m *Manager) check() (ok bool) {
	select {
	case <-m.stop:
		m.conns = nil
		ok = true
	default:
		ok = false
	}
	return
}

func (m *Manager) Stop() {
	for _, conn := range m.conns {
		conn.close()
	}
	close(m.stop)
	close(m.postman)
	close(m.metaCh)
}

func newConn(conn net.Conn) *Conn {
	return &Conn{
		conn:   conn,
		reader: make(chan []byte),
		stop:   make(chan struct{}),
	}
}

func (c *Conn) start(postman chan<- []byte) {
	go c.read()
	go c.send(postman)
}

func (c *Conn) send(receiver chan<- []byte) {
	for bytes := range c.reader {
		if c.check() {
			break
		}
		receiver <- bytes
	}
}

func (c *Conn) read() {
	for {
		if c.check() {
			break
		}
		var buf []byte
		_ = c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, err := c.conn.Read(buf)
		if err != nil {
			log.Errorf("read data from conn(%s) failed, err: %+v", c.addr(), err)
			continue
		}
		c.reader <- buf
	}
}

func (c *Conn) write(buf []byte) {
	go func(buf []byte) {
		if c.check() {
			return
		}
		_ = c.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		_, err := c.conn.Write(buf)
		if err != nil {
			log.Errorf("write date to conn(%s) failed, err: %+v", c.addr(), err)
		}
	}(buf)
}

func (c *Conn) addr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Conn) check() (ok bool) {
	select {
	case <-c.stop:
		ok = true
	default:
		ok = false
	}
	return ok
}

func (c *Conn) close() {
	close(c.stop)
	close(c.reader)
	_ = c.conn.Close()
}
