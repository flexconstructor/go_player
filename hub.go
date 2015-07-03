// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package  go_player

import(

	player_log "github.com/flexconstructor/go_player/log"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {

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

	}
}

func (h *hub) run() {

	h.log.Info("Hub run: url = %s id= %s",h.stream_url,h.stream_id)

	ff=&ffmpeg{
		stream_url: h.stream_url+"/"+h.stream_id+"?model_id="+h.stream_id+"&access_token="+h.service_token,
		broadcast:h.broadcast,
		close_channel: make(chan bool),
		metadata: h.metadata,
		error: h.error,
		log: h.log,
		workers_length:20,

	}

	defer ff.Close();

	h.log.Debug("decoder created")

	/*conn = &RtmpConnector{
 		rtmp_url:	h.stream_url,
 		stream_id: h.stream_id,
		error_cannel: h.error,
		close_channel:make(chan bool),
		log: h.log,
		service_token: h.service_token,
 		 handler: &RtmpHandler{
 			 stream_status: h.rtmp_status,
			  error_channel: h.error,
			 log: h.log,
 		 },
	}
	defer conn.Close()
	h.log.Debug("connection created")
	*/
	go ff.run()
	h.log.Debug("run decoder")
	//defer h.Close()

	for {
		select {
		case c, ok := <-h.register:
		if(! ok){
			return
		}
			if(len(h.connections)==0){
				h.log.Debug("first connection")
				//go conn.Run()

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

				delete(h.connections, c)
				c=nil
				h.log.Debug("unregister connection. connection length: %d", len(h.connections))
				if (len(h.connections)==0) {
					return
				}
			}else{
				return
			}
		case m, ok := <-h.broadcast:
			if(! ok){
				h.log.Error("can not write chunk!")
				return
			}
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

		/*case s, o := <- h.rtmp_status:
			h.log.Debug("RTMP STATUS: %g",s)
			if(s==0) {
			h.log.Debug(">>>>>>Close rtmp")
			return
			}else{

			}
		*/
		case meta, ok := <- h.metadata:
		if(! ok){
			return
		}
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
		case e, ok:= <-h.error:
		if(! ok){
			return
		}
		h.log.Error("player error: %s",e.description)

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
	h.log.Debug("Close hub %s",h.stream_id)

	/*if(len(h.connections)>0){
		h.log.Debug("close connections %d",len(h.connections))
		for c := range h.connections {
			c.Close()
		}


	}
	h.exit_channel <- h
	h.log.Debug("hub closed")*/
}