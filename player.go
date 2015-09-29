package go_player

import (
	"errors"
	player_log "github.com/flexconstructor/go_player/log"

	"fmt"
	"runtime"
	"time"
)

/*
	RTMP-stream to JPEG-stream convertor.
*/
var (
	player_instance *GoPlayer         = nil //singleton instance
	log             player_log.Logger       // logger
)

type GoPlayer struct {
	streams_map   map[string]*hub    // map of stream connections hub.
	connects      chan *WSConnection // channel for register new web-socket connection.
	updates       chan *WSConnection // channel for updates connection.
	closes        chan *WSConnection // channel for unregister connections.
	stops         chan *GoPlayer     // close convertor instance channel.
	log           player_log.Logger  // logger.
	handler       IConnectionHandler // handler of connection events.
	broadcast_map map[uint64]*hub
	start_time    int64
}

// Init new player instance.
func InitGoPlayer(
	log player_log.Logger,
	connectionHandler IConnectionHandler) *GoPlayer {

	if player_instance != nil {
		return player_instance
	}
	player_instance = &GoPlayer{

		log:           log,
		streams_map:   make(map[string]*hub),
		connects:      make(chan *WSConnection, 1),
		updates:       make(chan *WSConnection, 1),
		closes:        make(chan *WSConnection, 1),
		stops:         make(chan *GoPlayer, 1),
		handler:       connectionHandler,
		broadcast_map: make(map[uint64]*hub),
		start_time:    time.Now().Unix(),
	}
	player_instance.log.Info("init go player")
	return player_instance
}

// return instance of player.
func GetPlayerInstance() (*GoPlayer, error) {
	if player_instance == nil {
		return nil, errors.New("goplayer not initialized!")
	}
	return player_instance, nil
}

// run the player instance.
func (p *GoPlayer) Run() {
	p.log.Info("Run GO PLAYER INSTANCE")
	defer p.stopInstance()
	defer p.recoverPlayer()
	for {
		select {
		// stop player instance
		case <-p.stops:
			return
		// init new connection
		case c, ok := <-p.connects:
			if ok {
				p.initConnection(c)
			} else {
				p.log.Error("can not write connection")
			}
		// close connection
		case c, ok := <-p.closes:
			if ok {
				p.closeConnection(c)

			} else {
				p.log.Error("can not close connection")
			}
		// update connection
		case u, ok := <-p.updates:
			if !ok {
				p.log.Debug("can not write update")
			}
			err := p.handler.OnUpdate(u)
			if err != nil {
				p.log.Debug("Update failed")
				u.error_channel <- err
			}

		default:
			if len(p.streams_map) > 0 {
				for i := range p.streams_map {
					h := p.streams_map[i]
					if h != nil {
						if len(h.connections) == 0 {
							p.log.Debug("call close hub")
							h.Close()
							p.log.Debug("hub closed")
							delete(p.streams_map, h.stream_url)
							p.log.Debug("hub deleted")
						}
					}

				}

			}

		}
	}

}

// Stop player.
func (p *GoPlayer) Stop() {
	p.log.Info("STOP GO PLAYER INSTANCE")
	p.stops <- p
}

func (p *GoPlayer) InitStream(stream_id uint64, stream_url string) {
	fmt.Println("init stream: %b url: %s", stream_id, stream_url)
	_, ok := p.broadcast_map[stream_id]
	if !ok {

		h := NewHub(stream_url, p.log, int(time.Now().Unix()-p.start_time), stream_id)
		p.broadcast_map[stream_id] = h
		go h.run()
	}
	return
}

func (p *GoPlayer) CloseStream(stream_id uint64) {
	fmt.Println("Close stream stream: %b", stream_id)
	h, ok := p.broadcast_map[stream_id]
	if !ok {
		p.log.Error("Can not finde stream %b for close!", stream_id)
		fmt.Println("Can not finde stream %b for close!", stream_id)
	} else {
		h.Close()
		delete(p.broadcast_map, stream_id)
	}
	return
}

// Stop player instance.
func (p *GoPlayer) stopInstance() {
	p.log.Info("Player stopped")
	player_instance = nil
}

// Register new web-socket connection.
func (p *GoPlayer) initConnection(conn *WSConnection) {
	fmt.Println("connect to: %b", conn.streamID)
	h, ok := p.broadcast_map[conn.streamID]
	if !ok {
		p.log.Error("hub for stream %b not found!", conn.streamID)
		conn.Close()
		return
	}
	h.register <- conn
	err := p.handler.OnConnect(conn)
	if err != nil {
		conn.error_channel <- err
	}
}

// Register close connection.
func (p *GoPlayer) closeConnection(conn *WSConnection) {
	fmt.Println("close connection to: %b", conn.streamID)
	h, ok := p.broadcast_map[conn.streamID]
	if !ok {
		p.log.Error("hub for stream %b not found!", conn.streamID)
		return
	}
	h.unregister <- conn
	err := p.handler.OnDisconnect(conn)
	if err != nil {
		p.log.Error("disconnection error %s", err.description)
	} else {
		return
	}
}

// recover player
func (p *GoPlayer) recoverPlayer() {
	if r := recover(); r != nil {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, false)
		reason := fmt.Sprintf("%v: %s", r, buf)
		p.log.Error("Runtime failure, reason -> %s", reason)
	}
}
