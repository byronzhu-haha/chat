package conn

import (
	"github.com/byronzhu-haha/chat/client/config"
	"github.com/byronzhu-haha/log"
	"net"
	"time"
)

type Conn struct {
	conn   net.Conn
	reader chan []byte
	stopCh chan struct{}
}

func NewConn() *Conn {
	return &Conn{
		reader: make(chan []byte, 1000),
		stopCh: make(chan struct{}),
	}
}

func (c *Conn) Start() error {
	conn, err := net.DialTimeout("tcp", config.DefaultConfig.ServerAddr, time.Second*3)
	if err != nil {
		return err
	}
	c.conn = conn
	c.work()
	return nil
}

func (c *Conn) work() {
	go func() {
		for {
			select {
			case <-c.stopCh:
				log.Infof("stop read data...")
				return
			default:
				var data []byte
				_, err := c.conn.Read(data)
				if err != nil {
					log.Errorf("read data failed, err: %+v", err)
					continue
				}
				c.reader <- data
			}
		}
	}()
}

func (c *Conn) ReceiveMsg() <-chan []byte {
	out := make(chan []byte, 1000)
	go func() {
		defer close(out)
		for bytes := range c.reader {
			select {
			case <-c.stopCh:
				return
			default:
				out <- bytes
			}
		}
	}()

	return out
}

func (c *Conn) SendMsg(msg []byte) {
	go func() {
		select {
		case <-c.stopCh:
			return
		default:
			_, err := c.conn.Write(msg)
			if err != nil {
				log.Errorf("send message failed, err: %+v", err)
				return
			}
		}
	}()
}

func (c *Conn) Stop() {
	_ = c.conn.Close()
	close(c.stopCh)
}
