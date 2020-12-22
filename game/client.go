package game

import (
	"context"
	"log"
	"time"

	"nhooyr.io/websocket"
)

const (
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 5 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	// Inbound messages from the server.
	Recv chan []byte

	// Outbound messages to the server.
	Send chan []byte

	// Disconnect
	Disconnect chan bool
	conn       *websocket.Conn
	Latency    int64
}

func NewClient() *Client {
	return &Client{
		Recv:       make(chan []byte, 256),
		Disconnect: make(chan bool),
		conn:       nil,
		Send:       make(chan []byte, 256),
	}
}

func (c *Client) Dial(addr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws://"+addr+"/ws", nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) DialTLS(addr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "wss://"+addr+"/ws", nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Listen(ctx context.Context) {
	go c.writePump(ctx)
	go c.readPump(ctx)
}

func (c *Client) SendMessage(message []byte) {
	c.Send <- message
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.conn.Close(websocket.StatusInternalError, "unexpected close")
		c.Disconnect <- true
	}()
	c.conn.SetReadLimit(maxMessageSize)
	for {
		_, buf, err := c.conn.Read(ctx)
		if err != nil {
			break
		}

		c.Recv <- buf
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusInternalError, "unexpected close")
		c.Disconnect <- true
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel.
				c.conn.Write(ctx, websocket.MessageText, []byte("closed"))
				log.Println("server closed")
				return
			}

			w, err := c.conn.Writer(ctx, websocket.MessageBinary)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			lastPinged := time.Now()
			err := c.conn.Ping(ctx)
			if err != nil {
				c.Latency = 999
				return
			}
			c.Latency = time.Since(lastPinged).Milliseconds()
		}
	}
}
