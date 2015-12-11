package go_player

import (
	player_log "github.com/flexconstructor/go_player/log"
	"fmt"
	"net"
	"os"
	"strconv"
)

/* The pool of web-socket connections for one model-stream.
   The struct contains, map of web-socket connections for this.stream,
   metadata object of this stream.
*/
type hub struct {
	connections  map[*WSConnection]bool // Registered connections.
	broadcast    chan []byte            // Channel for jpeg stream for client.
	register     chan *WSConnection     // Channel for register new connection.
	unregister   chan *WSConnection     // Channel for unregister connection.
	exit_channel chan bool              // Channel for close hub.
	metadata     chan *MetaData         // Stream metadata chennel.
	error        chan *WSError          // Error channel.
	log          player_log.Logger      // logger instance
	model_id     uint64                 // Stream id
	listener     net.Listener           // unix socket listener.
	socket_dir   string                 // directory for unix.socket file
}

var meta *MetaData // metadata of stream

// Create new hub instance.
func NewHub(
logger player_log.Logger, model_id uint64, socket_dir string,
) *hub {
	return &hub{
		broadcast:    make(chan []byte),
		register:     make(chan *WSConnection),
		unregister:   make(chan *WSConnection),
		connections:  make(map[*WSConnection]bool),
		metadata:     make(chan *MetaData),
		error:        make(chan *WSError, 1),
		log:          logger,
		exit_channel: make(chan bool, 100),
		model_id:     model_id,
		socket_dir:   socket_dir,
	}
}

// Run hub instance.
func (h *hub) run() {
	sock := fmt.Sprintf("%s/%s.sock", h.socket_dir, strconv.FormatUint(h.model_id, 10))
	fileinfo, err := os.Stat(sock)
	if err == nil && fileinfo != nil {
		err = os.Remove(sock)
		if err != nil {
			h.log.Error("can not remove existed socket: %s", err)
		}
	}
	go h.listenSocket(sock)
	defer h.closeSocketConnection(sock)
	for {
		select {
		case <-h.exit_channel:
			return
		case c, ok := <-h.register:
			if !ok {
				continue
			}
			h.connections[c] = true
		// try send metadata of stream to client.
			if meta != nil {
				b, err := meta.JSON()
				if err == nil {
					c.metadata <- b
				}
			}
		// unregister web-socket connection when connection been closed.
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
			} else {
				h.log.Debug("attempt duplicate disconnect")
				continue
			}
		// send new jpeg data for clients.
		case m, ok := <-h.broadcast:
			if !ok {
				continue
			}
			if len(h.connections) > 0 {
				for c := range h.connections {
					if c.HasVideo() == 1 {
						c.send <- m
					}
				}
			} else {
				continue
			}
		// close all connection when error message income
		case e, ok := <-h.error:
			if !ok {
				continue
			}
			for c := range h.connections {
				select {
				case c.error_channel <- e:
				default:
					delete(h.connections, c)
				}
			}
		}
	}
}

// Listen unix socket.
func (h *hub) listenSocket(socket_path string) {
	l, err := net.Listen("unix", socket_path)
	if err != nil {
		h.log.Error("listen error: %s", err)
		return
	}
	h.listener = l
	for {
		fd, err := h.listener.Accept()
		if err != nil {
			h.log.Error("accept error: %s", err)
			return
		}
		go h.echoServer(fd)
	}
}

// Make complete jpeg picture and send to
// web-socket connection.
func (h *hub) echoServer(c net.Conn) {
	total_buffer := make([]byte, 0)
	defer func() {
		h.broadcast <- total_buffer
		c.Close()
	}()
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}
		data := buf[0:nr]
		total_buffer = append(total_buffer, data...)
	}
}

// Close unix socket.
func (h *hub) closeSocketConnection(unix_file_path string) {
	if h.listener == nil {
		return
	}
	err := h.listener.Close()
	if err != nil {
		h.log.Error("Can not close connection %v", err)
	}
	error := os.Remove(unix_file_path)
	if error != nil {
		h.log.Error("can not remove unix socket file %v", error)
	}
	h.listener = nil
}

// close hub function
func (h *hub) Close() {
	h.exit_channel <- true
}