package client

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5000 * time.Millisecond

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	// Inbound messages from the server.
	recv chan []byte

	// Outbound messages to the server.
	send chan []byte

	// Disconnect
	disconnect chan bool
	conn       *websocket.Conn
}

func NewClient() *Client {
	return &Client{
		recv:       make(chan []byte, 256),
		disconnect: make(chan bool),
		conn:       nil,
		send:       make(chan []byte, 256),
	}
}

func (c *Client) Dial(addr string) error {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) DialTLS(addr string) error {
	conn, _, err := websocket.DefaultDialer.Dial("wss://"+addr+"/ws", nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Listen() {
	go c.writePump()
	go c.readPump()
}

func (c *Client) ChRecv() chan []byte { return c.recv }

func (c *Client) ChDisconnected() chan bool { return c.disconnect }

func (c *Client) SendMessage(message []byte) {
	c.send <- message
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.disconnect <- true
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, buf, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.recv <- buf
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.disconnect <- true
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
