package go_player

import (
	player_log "github.com/flexconstructor/go_player/log"

	"fmt"
	"runtime"
	"net"
	"strconv"
	"os"

)

/* The pool of web-socket connections for one model-stream.
   The struct conains decode/encode module for one rtmp-stream,
   map of web-socket connections for this.stream,
   metadata object of this stream.
*/
type hub struct {
	stream_url   string                 // RTMP stream url.
	connections  map[*WSConnection]bool // Registered connections.
	broadcast    chan []byte            // Channel for jpeg stream for client.
	register     chan *WSConnection     // Channel for register new connection.
	unregister   chan *WSConnection     // Channel for unregister connection.
	exit_channel chan bool              // Channel for close hub.
	metadata     chan *MetaData         // Stream metadata chennel.
	error        chan *WSError          // Error channel.
	log          player_log.Logger      // Logger reference.
	hub_id       int
	model_id     uint64
}

var ff *ffmpeg     // FFMPEG module for decode/encode stream
var meta *MetaData // metadata of stream
var _listener net.Listener

// Create new hub instance.
func NewHub(stream_url string,
	logger player_log.Logger, hub_id int, model_id uint64,
) *hub {
	return &hub{
		stream_url:   stream_url,
		broadcast:    make(chan []byte),
		register:     make(chan *WSConnection),
		unregister:   make(chan *WSConnection),
		connections:  make(map[*WSConnection]bool),
		metadata:     make(chan *MetaData),
		error:        make(chan *WSError, 1),
		log:          logger,
		exit_channel: make(chan bool, 100),
		hub_id:       hub_id,
		model_id:    model_id,
	}
}

// run hub instance.
func (h *hub) run() {
	/*h.log.Info("Hub run: url = %s ", h.stream_url)
	fmt.Println("Hub run: url = %s ", h.stream_url)
	ff = &ffmpeg{
		stream_url:     h.stream_url,
		broadcast:      h.broadcast,
		close_channel:  make(chan bool),
		metadata:       h.metadata,
		error:          h.error,
		log:            h.log,
		workers_length: 1,
		hub_id:         h.hub_id,
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
			//h.connections[0].send <-m
			if len(h.connections) > 0 {
				for c := range h.connections {
					c.send <- m
				}
			} else {
				continue
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
	}*/
	defer h.recoverHub()


	sock:=fmt.Sprintf("/home/mediaapi/nginx/html/temp/dash/%s.sock",strconv.FormatUint(h.model_id,10))
	go h.listenSocket(sock)
	//l, err := net.Listen("unix", sock)

	//if err != nil {
		//log.Fatal("listen error:", err)
		//fmt.Println("listen error: %s",err)
	//}
	defer closeSocketConnection(sock)
	/*for {
		select {
		 case <- h.exit_channel:
		 fmt.Println("hub exit command income")
		return
		default:

		}

	}*/

		for{
			select{
			case <- h.exit_channel:
				fmt.Println("hub exit command income")
			return
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
				fmt.Printf("jpg data: %v\n",len(m))
			//h.connections[0].send <-m
				if len(h.connections) > 0 {
					for c := range h.connections {
						c.send <- m
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

func (h *hub)listenSocket(socket_path string){
l, err:= net.Listen("unix", socket_path)

	if err != nil {
		fmt.Println("listen error: %s",err)
		return
		}
	_listener=l
	for {
		fd, err := _listener.Accept()
		if err != nil {
			fmt.Println("accept error: %s", err)
			return
		}
		go h.echoServer(fd)
	}
}

func (h *hub)echoServer(c net.Conn) {

	total_buffer:=make([]byte,0)

	defer func (){
		fmt.Printf("total bytes %v\n",len(total_buffer))
		h.broadcast <- total_buffer
		c.Close()
	}()

	for {
		buf := make([]byte,512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}
		data := buf[0:nr]
		total_buffer=append(total_buffer,data...)
	}

}

func closeSocketConnection(unix_file_path string){
	if(_listener==nil){
		return
	}
	defer fmt.Println("close socket")
	err:= _listener.Close()
	if(err != nil){
		fmt.Errorf("Can not close connection %v",err)
	}

	 error:= os.Remove(unix_file_path)

	if(error!= nil){
		fmt.Errorf("can not remove unix socket file %v",error)
	}
	_listener=nil

}


// close hub function
func (h *hub) Close() {
	h.log.Debug("Close hub %s", h.stream_url)
	fmt.Println("Close!!! hub %s", h.stream_url)
	h.exit_channel <- true
	fmt.Println("write to close channel")
}

func (h *hub) recoverHub() {
	if r := recover(); r != nil {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, false)
		reason := fmt.Sprintf("%v: %s", r, buf)
		h.log.Error("Runtime failure, reason -> %s", reason)
	}
}
