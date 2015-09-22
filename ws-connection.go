package go_player

import (
	player_log "github.com/flexconstructor/go_player/log"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"runtime"
	"fmt"
)
/*
  Web-socket connection instance.
 */
const (
	// Time allowed to write a message to the peer.
	writeWait = 2 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 200 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 1) / 2

	// Maximum message queue allowed from/to peer.
	maxMessageSize = 512
)

// web-socket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}



type WSConnection struct {
	ws            *websocket.Conn
	send          chan []byte
	metadata      chan []byte
	error_channel chan *WSError
	lgr           player_log.Logger
	request       *http.Request
	source_url    string
}

// Create new web-socket instance.
func NewWSConnection(source_url string, w http.ResponseWriter, r *http.Request, l player_log.Logger) (*WSConnection, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	conn := &WSConnection{
		source_url: source_url,
		ws:            ws,
		send:          make(chan []byte, 256),
		error_channel: make(chan *WSError, 1),
		metadata:      make(chan []byte),
		lgr:           l,
		request:       r,
	}

	return conn, nil
}

// Write message to web-socket buffer.
func (c *WSConnection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// Run new web-socket connection.
func (c *WSConnection) Run() {
	player, err := GetPlayerInstance()
	if err != nil {
		c.lgr.Error("no player instance found")
		return
	}
	player.connects <- c
	c.readPump()
	fmt.Println("after pump")
	ticker := time.NewTicker(pingPeriod)
	defer c.Close()
	defer ticker.Stop()
	for {
		select {
		// send message.
		case message, ok := <-c.send:
			if !ok {
				c.lgr.Error("can not write message")
				return
			}

			if err := c.write(websocket.BinaryMessage, message); err != nil {
				c.lgr.Error("can not wright binary")
				return
			}
		// send ping.
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				c.lgr.Error("can not write ping")
				fmt.Println("can not write ping")
				return
			}
			err := c.callUpdate()
			if err != nil {
				c.lgr.Error("Update error")
				return
			}
		fmt.Println("ping")
		// send metadata.
		case metadata, ok := <-c.metadata:
			if ok {
				c.write(websocket.TextMessage, metadata)
			}
		// send error message.
		case error, ok := <-c.error_channel:
			if ok {
				error_object, err := error.JSON()
				if err == nil {
					c.write(websocket.TextMessage, error_object)
				}
				// close connection if error is critical.
				if error.level == 1 {
					c.lgr.Error("error level = %d descripton= %s", error.level, error.description)
					return
				}
			}
		}
	}

}

// readPump pumps messages from the websocket connection to the hub.
func (c *WSConnection) readPump() {
	fmt.Println("readPump>>>")
	defer func() {
		//c.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		fmt.Println("pong")
		return nil

	})
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		fmt.Println("message: %s",message)
	}
}

// close connection.
func (c *WSConnection) Close() {
	write_error := c.write(websocket.CloseMessage, []byte{})
	if write_error != nil {
		c.lgr.Error("can not write close message")

	}
	c.ws.Close()
	fmt.Println("close connection")
	pl, err := GetPlayerInstance()
	if err != nil {
		c.lgr.Error("NO Player found: ", err)
		return
	}
	pl.closes <- c
}

// dispatch update event for handler.
func (c *WSConnection) callUpdate() error {
	pl, err := GetPlayerInstance()
	if err != nil {
		c.lgr.Error("NO Player found: ", err)
		return err
	}
	pl.updates <- c
	return nil
}

// Get http.Request instance for public.
func (c *WSConnection)GetRequest()(* http.Request){
	return c.request
}
// Get connection url for public.
func (c *WSConnection)GetSourceURL()(string){
	return c.source_url
}


func(c *WSConnection) recoverConnection() {
	if r := recover(); r != nil {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, false)
		reason := fmt.Sprintf("%v: %s", r, buf)
		c.lgr.Error("Runtime failure, reason -> %s", reason)
	}
}
