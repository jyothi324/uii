package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 4
)

type Client struct {
	manager   *Manager
	conn      *gorillawebsocket.Conn
	send      chan []byte
	ctx       context.Context
	tenant    string
	cancel    context.CancelFunc
	closeOnce sync.Once
}

func newClient(conn *gorillawebsocket.Conn, manager *Manager, tenant string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		manager: manager,
		conn:    conn,
		tenant:  tenant,
		send:    make(chan []byte, 256),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (c *Client) start() {
	c.manager.Register(c)
	go c.writePump()
	go c.readPump()
}

func (c *Client) remoteAddr() string {
	if c.conn == nil || c.conn.RemoteAddr() == nil {
		return "unknown"
	}
	return c.conn.RemoteAddr().String()
}

func (c *Client) readPump() {
	defer func() {
		c.manager.Unregister(c)
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if gorillawebsocket.IsUnexpectedCloseError(err, gorillawebsocket.CloseGoingAway, gorillawebsocket.CloseAbnormalClosure) {
				log.Printf("websocket read error from %s: %v", c.remoteAddr(), err)
			}
			return
		}
		// Server-push only: ignore payloads from the client.
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(gorillawebsocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(gorillawebsocket.TextMessage)
			if err != nil {
				log.Printf("websocket next writer error for %s: %v", c.remoteAddr(), err)
				return
			}
			if _, err := writer.Write(message); err != nil {
				log.Printf("websocket write error for %s: %v", c.remoteAddr(), err)
				_ = writer.Close()
				return
			}
			if err := writer.Close(); err != nil {
				log.Printf("websocket writer close error for %s: %v", c.remoteAddr(), err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteControl(gorillawebsocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				log.Printf("websocket ping error for %s: %v", c.remoteAddr(), err)
				return
			}
		}
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		c.cancel()
		_ = c.conn.Close()
	})
}
