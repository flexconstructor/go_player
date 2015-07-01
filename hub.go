// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package  go_player

import(
	"sync"
	player_log "github.com/flexconstructor/go_player/log"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	sync.Mutex
	//url of stream application
	stream_url string
	//stream name
	stream_id string
	// Registered connections.
	connections map[*WSConnection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *WSConnection

	// Unregister requests from connections.
	unregister chan *WSConnection

	rtmp_status chan int
	exit_channel chan *hub
	rtmp_close chan bool
	ffmpeg_close chan bool
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
stream_name string,
logger player_log.Logger,
service_token string,
exit_channel chan *hub,
) *hub{
	return &hub{
		stream_url: stream_url,
		stream_id: stream_name,
		broadcast:   make(chan []byte),
		register:    make(chan *WSConnection),
		unregister:  make(chan *WSConnection),
		connections: make(map[*WSConnection]bool),
		rtmp_status: make(chan int, 256),
		metadata:  make(chan *MetaData),
		error: make(chan *WSError),
		log: logger,
		service_token: service_token,
		exit_channel: exit_channel,
		rtmp_close: make(chan bool),
		ffmpeg_close: make(chan bool),


	}
}

func (h *hub) run() {
	defer h.Close()
	h.log.Info("Hub run: url = %s id= %s",h.stream_url,h.stream_id)
	/*decoder=&FFmpegDecoder{
		stream_url: h.stream_url+"/"+h.stream_id,
		broadcast:h.broadcast,
		rtmp_status: h.rtmp_status,
		metadata: h.metadata,
		error: h.error,
		log: h.log,
	}*/
	ff=&ffmpeg{
		stream_url: h.stream_url+"/"+h.stream_id,
		broadcast:h.broadcast,
		rtmp_status: h.rtmp_status,
		metadata: h.metadata,
		error: h.error,
		log: h.log,
		close_chan: h.ffmpeg_close,
		workers_length:20,

	}
	h.log.Debug("decoder created")

	conn = &RtmpConnector{
 		rtmp_url:	h.stream_url,
 		stream_id: h.stream_id,
		error_cannel: h.error,
		close_channel:h.rtmp_close,
		log: h.log,
 		 handler: &RtmpHandler{
 			 stream_status: h.rtmp_status,
			  error_channel: h.error,
			 log: h.log,
 		 },
	}
	h.log.Debug("connection created")

	for {
		select {
		case c := <-h.register:
			if(len(h.connections)==0){
				h.log.Debug("first connection")
				go conn.Run()

			}
			h.connections[c] = true

		h.log.Debug("Register connection")

		if(meta != nil){
			b, err:=meta.JSON()
			if(err==nil) {
				c.metadata <- b
				h.log.Debug("send metadata")
			}
		}


		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				h.log.Debug("unregister connection")
				delete(h.connections, c)
				c=nil
				if(len(h.connections)==0){
					return
				}
			}
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					c.Close()
					delete(h.connections, c)
					c=nil
					if(len(h.connections)==0){
						return
					}
				}
			}

		case s := <- h.rtmp_status:
			if(s==0) {
			h.log.Debug("Close rtmp")
			return
			}else{
				go ff.run()
				h.log.Debug("run decoder")
			}
		h.log.Debug("RTMP STATUS: %g",s)
		case meta= <- h.metadata:
		b, err:=meta.JSON()
		if(err != nil){
			continue
		}
		h.log.Debug("new metadata")
			for c := range h.connections {
				select {
				case c.metadata <- b:
				default:
					c.Close()
					delete(h.connections, c)
					c=nil
					if(len(h.connections)==0){
						return
					}
				}
			}
		case e:= <-h.error:
		h.log.Error("player error",e.description)
			for c := range h.connections {
				select {
				case c.error_channel <- e:
				default:
					c.Close()
					delete(h.connections, c)
				    c=nil
					if(len(h.connections)==0){
						return
					}
				}
			}

		}
	}


}


func (h *hub)Close(){
	h.log.Debug("Close hub ",h.stream_id)

	for c := range h.connections {
		c.Close()
	}
	h.exit_channel <- h
	h.ffmpeg_close <- true
	h.rtmp_close <- true

	h.log.Debug("hub closed")

}