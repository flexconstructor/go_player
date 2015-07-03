// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package  go_player

import(

	player_log "github.com/flexconstructor/go_player/log"
	"strconv"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {

	//url of stream application
	stream_url string
	//stream name
	stream_id uint64
	// Registered connections.
	connections map[*WSConnection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *WSConnection

	// Unregister requests from connections.
	unregister chan *WSConnection

	rtmp_status chan int
	exit_channel chan bool

	metadata chan *MetaData
	error chan *WSError
	log player_log.Logger
	service_token string
	//connection_handler IConnectionHandler
}

//var decoder *FFmpegDecoder
var ff *ffmpeg
var conn *RtmpConnector
var meta *MetaData


func NewHub(stream_url string,
stream_id uint64,
logger player_log.Logger,
service_token string,
) *hub{
	return &hub{
		stream_url: stream_url,
		stream_id: stream_id,
		broadcast:   make(chan []byte),
		register:    make(chan *WSConnection),
		unregister:  make(chan *WSConnection),
		connections: make(map[*WSConnection]bool),
		rtmp_status: make(chan int, 256),
		metadata:  make(chan *MetaData),
		error: make(chan *WSError),
		log: logger,
		service_token: service_token,
		exit_channel: make(chan bool),
	}
}

func (h *hub) run() {
	h.log.Info("Hub run: url = %s id= %s",h.stream_url,h.stream_id)
	stream_name:=strconv.FormatUint(h.stream_id,10)
	ff=&ffmpeg{
		stream_url: h.stream_url+"/"+stream_name+"?model_id="+stream_name+"&access_token="+h.service_token,
		broadcast:h.broadcast,
		close_channel: make(chan bool),
		metadata: h.metadata,
		error: h.error,
		log: h.log,
		workers_length:256,
	}

	go ff.run()
	defer ff.Close();

	for {
		select {
		case c, ok := <-h.register:
		if(! ok){
			continue
		}

			h.connections[c] = true
			h.log.Debug("register connection: %d",len(h.connections))

		if(meta != nil){
			b, err:=meta.JSON()
			if(err==nil) {
				c.metadata <- b

			}
		}
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				h.log.Debug("unregister connection. connection length: %d", len(h.connections))

			}else{
				h.log.Error("can not unregister connection")
				continue
			}
		case m, ok := <-h.broadcast:
			if(! ok){
				h.log.Error("can not write chunk!")
				continue
			}
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					c.Close()
					delete(h.connections, c)

				}
			}

		case meta, ok := <- h.metadata:
		if(! ok){
			continue
		}
		b, err:=meta.JSON()
		if(err != nil){
			continue
		}

			for c := range h.connections {
				select {
				case c.metadata <- b:
				default:
					c.Close()
					delete(h.connections, c)

				}
			}
		case e, ok:= <-h.error:
		if(! ok){
			continue
		}
			for c := range h.connections {
				select {
				case c.error_channel <- e:
				default:
					c.Close()
					delete(h.connections, c)
				}
			}
		case <-h.exit_channel:
		return

		}
	}
}


func (h *hub)Close(){
	h.log.Debug("Close hub %s",h.stream_id)
	h.exit_channel <- true

}