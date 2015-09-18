package go_player

import (
	"errors"
	player_log "github.com/flexconstructor/go_player/log"

)
/*
	RTMP-stream to JPEG-stream convertor.
 */
var (
	player_instance *GoPlayer = nil    //singleton instance
	log             player_log.Logger  // logger
)

type GoPlayer struct {

	streams_map   map[string]*hub     // map of stream connections hub.
	connects      chan *WSConnection  // channel for register new web-socket connection.
	updates       chan *WSConnection  // channel for updates connection.
	closes        chan *WSConnection  // channel for unregister connections.
	stops         chan *GoPlayer      // close convertor instance channel.
	log           player_log.Logger   // logger.
	handler       IConnectionHandler  // hundler of connection events.
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
							h.Close()
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

// Stop player instance.
func (p *GoPlayer) stopInstance() {
	p.log.Info("Player stopped")
	player_instance = nil
}

// Register new web-socket connection.
func (p *GoPlayer) initConnection(conn *WSConnection) {
	stream_url:=conn.GetSourceURL()
	//h, ok := p.streams_map[stream_url]
	// if hub of requested stream not running - run new hub.
	//if !ok {
		h:= NewHub(
			stream_url,
			p.log,
		)

		p.streams_map[stream_url] = h
		p.log.Debug("NEW STREAM: %d",len(p.streams_map))
		go h.run()
	//}
	// register connection in hub.
	h.register <- conn
	err := p.handler.OnConnect(conn)
	if err != nil {
		conn.error_channel <- err
	}

}

// Register close connection.
func (p *GoPlayer) closeConnection(conn *WSConnection) {
	stream_url:= conn.GetSourceURL();

	p.log.Debug("Close connection with params:source url=  %s", stream_url)
	h, ok := p.streams_map[stream_url]
	if !ok {
		p.log.Error("hub for stream %s not found!", stream_url)
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
