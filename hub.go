// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go_player

import (
	player_log "github.com/flexconstructor/go_player/log"

)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {

	//url of stream application
	stream_url string
	// Registered connections.
	connections map[*WSConnection]bool
	// Inbound messages from the connections.
	broadcast chan []byte
	// Register requests from the connections.
	register chan *WSConnection
	// Unregister requests from connections.
	unregister    chan *WSConnection
	exit_channel  chan bool
	metadata      chan *MetaData
	error         chan *WSError
	log           player_log.Logger

}

//var decoder *FFmpegDecoder
var ff *ffmpeg
var meta *MetaData

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

func (h *hub) run() {
	h.log.Info("Hub run: url = %s ", h.stream_url)
	ff = &ffmpeg{
		stream_url:     h.stream_url,
		broadcast:      h.broadcast,
		close_channel:  make(chan bool),
		metadata:       h.metadata,
		error:          h.error,
		log:            h.log,
		workers_length: 256,
	}

	go ff.run()
	defer ff.Close()

	for {
		select {
		case c, ok := <-h.register:
			if !ok {
				continue
			}
			h.connections[c] = true
			h.log.Debug("register connection: %d", len(h.connections))

			if meta != nil {
				b, err := meta.JSON()
				if err == nil {
					c.metadata <- b

				}
			}
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				h.log.Debug("unregister connection. connection length: %d", len(h.connections))

			} else {
				continue
			}
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
		case <-h.exit_channel:
			return

		}
	}
}

func (h *hub) Close() {
	h.log.Debug("Close hub %s", h.stream_url)
	h.exit_channel <- true
}
