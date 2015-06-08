// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package  GoPlayer

import(
	"sync"
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
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	rtmp_status chan int

	metadata chan *MetaData
	error chan *Error
}

var decoder *FFmpegDecoder
var conn *RtmpConnector
var meta *MetaData


func NewHub(stream_url string,stream_name string) *hub{
	return &hub{
		stream_url: stream_url,
		stream_id: stream_name,
		broadcast:   make(chan []byte),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
		rtmp_status: make(chan int, 0),
		metadata:  make(chan *MetaData),
		error: make(chan *Error),
	}
}

func (h *hub) run() {

	decoder=&FFmpegDecoder{
		stream_url: h.stream_url+"/"+h.stream_id,
		broadcast:h.broadcast,
		rtmp_status: h.rtmp_status,
		metadata: h.metadata,
		error: h.error,
	}


	conn = &RtmpConnector{
 		rtmp_url:	h.stream_url,
 		stream_id: h.stream_id,
		error_cannel: h.error,
 		 handler: &RtmpHandler{
 			 stream_status: h.rtmp_status,
			  error_channel: h.error,
 		 },
	}

	go conn.Run()

	//go decoder.Run()

	defer h.CloseHub(decoder)

	for {
		select {
		case c := <-h.register:
			h.connections[c] = true

		if(meta != nil){
			b, err:=meta.JSON()
			if(err==nil) {
				c.metadata <- b
			}
		}


		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				c.Close()
				delete(h.connections, c)
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
					if(len(h.connections)==0){
						return
					}
				}
			}

		case s := <- h.rtmp_status:
			if(s==0) {
				return
			}else{
				go decoder.Run()
			}
		case meta= <- h.metadata:
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
					if(len(h.connections)==0){
						return
					}
				}
			}
		case e:= <-h.error:
			for c := range h.connections {
				select {
				case c.error_channel <- e:
				default:
					c.Close()
					delete(h.connections, c)
					if(len(h.connections)==0){
						return
					}
				}
			}
				if(e.Level==1){
					return
				}

		}
	}


}


func (h *hub)CloseHub(decoder *FFmpegDecoder){

	if(h.register != nil){
		close(h.register)
		h.register=nil
	}
	if(h.rtmp_status != nil){
		close(h.rtmp_status)
		h.rtmp_status=nil
	}
	if(h.unregister != nil){
		close(h.unregister)
		h.unregister=nil
	}

	if(h.metadata != nil){
		close(h.metadata)
		h.metadata=nil
	}
	meta=nil
	decoder.Close()
	conn.Close()
	delete(NewGoPlayer().streams_map,h.stream_id)

}
