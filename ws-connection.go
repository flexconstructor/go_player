package go_player

import (
	player_log "github.com/flexconstructor/go_player/log"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 2 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 2 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 1) / 2

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

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

func (c *WSConnection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *WSConnection) Run() {
	c.lgr.Debug("Run connection")
	player, err := GetPlayerInstance()
	if err != nil {
		c.lgr.Error("no player instance found")
		return
	}
	player.connects <- c

	ticker := time.NewTicker(pingPeriod)
	defer c.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.lgr.Error("can not write message")
				return
			}
			if err := c.write(websocket.BinaryMessage, message); err != nil {
				c.lgr.Error("can not wright binary")
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				c.lgr.Error("can not write ping")
				return
			}
			err := c.callUpdate()
			if err != nil {
				c.lgr.Error("Update error")
				return
			}
		case metadata, ok := <-c.metadata:
			if ok {
				c.write(websocket.TextMessage, metadata)
			}
		case error, ok := <-c.error_channel:
			if ok {
				error_object, err := error.JSON()
				if err == nil {
					c.write(websocket.TextMessage, error_object)
				}
				if error.level == 1 {
					c.lgr.Error("error level = %d descripton= %s", error.level, error.description)
					return
				}
			}
		}
	}

}

func (c *WSConnection) Close() {
	//c.lgr.Debug("connection closed for user: %d", c.params.ClientID)
	write_error := c.write(websocket.CloseMessage, []byte{})
	if write_error != nil {
		c.lgr.Error("can not write close message")

	}
	c.ws.Close()
	pl, err := GetPlayerInstance()
	if err != nil {
		c.lgr.Error("NO Player found: ", err)
		return
	}
	pl.closes <- c
}


func (c *WSConnection) callUpdate() error {
	pl, err := GetPlayerInstance()
	if err != nil {
		c.lgr.Error("NO Player found: ", err)
		return err
	}
	pl.updates <- c
	return nil
}

func (c *WSConnection)GetRequest()(* http.Request){
	return c.request
}

func (c *WSConnection)GetSourceURL()(string){
	return c.source_url
}
