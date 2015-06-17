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
	log Logger
	service_token string
	connection_handler IConnectionHandler
}

var decoder *FFmpegDecoder
var conn *RtmpConnector
var meta *MetaData


func NewHub(stream_url string,
stream_name string,
logger Logger,
service_token string,
connection_handler IConnectionHandler) *hub{
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
		log: logger,
		service_token: service_token,
		connection_handler: connection_handler,

	}
}

func (h *hub) run() {
	h.log.Info("Hub run: ",h.stream_url,"id: ",h.stream_id)
	decoder=&FFmpegDecoder{
		stream_url: h.stream_url+"/"+h.stream_id,
		broadcast:h.broadcast,
		rtmp_status: h.rtmp_status,
		metadata: h.metadata,
		error: h.error,
		log: h.log,
	}
	h.log.Debug("decoder created")

	conn = &RtmpConnector{
 		rtmp_url:	h.stream_url,
 		stream_id: h.stream_id,
		error_cannel: h.error,
		log: h.log,
 		 handler: &RtmpHandler{
 			 stream_status: h.rtmp_status,
			  error_channel: h.error,
			 log: h.log,
 		 },
	}
	h.log.Debug("connection created")

	//h.log.Debug("connection runing")
	//go decoder.Run()

	defer h.CloseHub(decoder)

	for {
		select {
		case c := <-h.register:
			if(len(h.connections)==0){
				h.log.Debug("Run rtmp connection!!!!")
				//go conn.Run()
				go decoder.Run()
			}
			h.connections[c] = true
			h.registerConnection(c)
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
				h.log.Debug("close connection")
				c.Close()
				delete(h.connections, c)
				c=nil
				/*if(len(h.connections)==0){
					return
				}*/
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
				//go decoder.Run()
				//h.log.Debug("run decoder")
			}
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
		h.log.Error("player error",e)
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
				if(e.Level==1){
					return
				}

		}
	}


}

func (h *hub)registerConnection(conn *connection){
h.log.Debug("REGISTER CONNECTION")
h.log.Debug("client_id: ",conn.client_id)
	h.log.Debug("access_token: ",conn.access_token)
	h.log.Debug("model_id: ",conn.model_id)
	/*err:=h.connection_handler.OnConnect(IConnection(conn))
	if(err!= nil){
		h.log.Error("callback error on connect: ",err)
		conn.error_channel<- err
	}*/
}


func (h *hub)closeConnection(conn *connection){
	h.log.Debug("CLOSE CONNECTION")
	h.log.Debug("client_id: ",conn.client_id)
	h.log.Debug("access_token: ",conn.access_token)
	h.log.Debug("model_id: ",conn.model_id)
	err:= h.connection_handler.OnDisconnect(IConnection(conn))
	if(err != nil){
		h.log.Error("callback error on disconnect: ",err)
		conn.error_channel <- err
	}

}


func (h *hub)CloseHub(decoder *FFmpegDecoder){
	h.log.Debug("close hub")
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
	player, err:= GetPlayerInstance();
	if(err!=nil){
		return
	}
	delete(player.streams_map,h.stream_id)

}
