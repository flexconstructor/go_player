package go_player

import (
	player_log "github.com/flexconstructor/go_player/log"

)

/* The pool of web-socket connections for one model-stream.
   The struct conains decode/encode module for one rtmp-stream,
   map of web-socket connections for this.stream,
   metadata object of this stream.
*/
type hub struct {

	stream_url string                   // RTMP stream url.
	connections map[*WSConnection]bool  // Registered connections.
	broadcast chan []byte               // Channel for jpeg stream for client.
	register chan *WSConnection         // Channel for register new connection.
	unregister    chan *WSConnection    // Channel for unregister connection.
	exit_channel  chan bool             // Channel for close hub.
	metadata      chan *MetaData        // Stream metadata chennel.
	error         chan *WSError         // Error channel.
	log           player_log.Logger     // Logger reference.

}


var ff *ffmpeg      // FFMPEG module for decode/encode stream
var meta *MetaData  // metadata of stream

// Create new hub instance.
func NewHub(stream_url string,
	logger player_log.Logger,
) *hub {
	return &hub{
		stream_url:    stream_url,
		broadcast:     make(chan []byte),
		register:      make(chan *WSConnection, 1),
		unregister:    make(chan *WSConnection, 1),
		connections:   make(map[*WSConnection]bool),
		metadata:      make(chan *MetaData),
		error:         make(chan *WSError, 1),
		log:           logger,
		exit_channel:  make(chan bool, 1),
	}
}

// run hub instance.
func (h *hub) run() {
	h.log.Info("Hub run: url = %s ", h.stream_url)
	ff = &ffmpeg{
		stream_url:     h.stream_url,
		broadcast:      h.broadcast,
		close_channel:  make(chan bool),
		metadata:       h.metadata,
		error:          h.error,
		log:            h.log,
		workers_length: 1,
	}

	// run ffmpeg module.
	go ff.run()
	defer ff.Close()

	for {
		select {
		// register new web-socket connection
		case c, ok := <-h.register:
			if !ok {
				continue
			}
			h.connections[c] = true
			h.log.Debug("register connection: %d", len(h.connections))
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
				h.log.Debug("unregister connection. connection length: %d", len(h.connections))

			} else {
				continue
			}
		// send new jpeg data for clients.
		case m, ok := <-h.broadcast:
			if !ok {
				continue
			}
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)

				}
			}
		// send methadata, when it income.
		case meta, ok := <-h.metadata:
			if !ok {
				continue
			}
			b, err := meta.JSON()
			if err != nil {
				continue
			}

			for c := range h.connections {
				select {
				case c.metadata <- b:
				default:
					delete(h.connections, c)

				}
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
		// close hub if exit message income.
		case <-h.exit_channel:
			return

		}
	}
}
// close hub function
func (h *hub) Close() {
	h.log.Debug("Close hub %s", h.stream_url)
	h.exit_channel <- true
}
