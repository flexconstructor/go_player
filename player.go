package go_player

import (
	"errors"
	player_log "github.com/flexconstructor/go_player/log"

)

var (
	player_instance *GoPlayer = nil
	log             player_log.Logger
)

type GoPlayer struct {

	streams_map   map[string]*hub
	connects      chan *WSConnection
	updates       chan *WSConnection
	closes        chan *WSConnection
	stops         chan *GoPlayer
	log           player_log.Logger
	handler       IConnectionHandler
}

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

func GetPlayerInstance() (*GoPlayer, error) {
	if player_instance == nil {
		return nil, errors.New("goplayer not initialized!")
	}
	return player_instance, nil
}

func (p *GoPlayer) Run() {
	p.log.Info("Run GO PLAYER INSTANCE")
	defer p.stopInstance()
	for {
		select {
		case <-p.stops:
			return
		case c, ok := <-p.connects:
			if ok {
				p.initConnection(c)
			} else {
				p.log.Error("can not write connection")
			}
		case c, ok := <-p.closes:
			if ok {
				p.closeConnection(c)

			} else {
				p.log.Error("can not close connection")
			}
		case u, ok := <-p.updates:
			if !ok {
				panic("can not write update")
			}
			err := p.handler.OnUpdate(u)
			if err != nil {
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

func (p *GoPlayer) Stop() {
	p.log.Info("STOP GO PLAYER INSTANCE")
	p.stops <- p
}

func (p *GoPlayer) stopInstance() {
	p.log.Info("Player stopped")
	player_instance = nil
}

func (p *GoPlayer) initConnection(conn *WSConnection) {
	//params := conn.params
	stream_url:=conn.GetSourceURL()
	p.log.Debug("connect to url: %s",stream_url)
	//p.log.Info("init connection  with params: stream_id=  %d user_id= %d access_token= %s", params.StreamID, params.ClientID, params.AccessToken)
	h, ok := p.streams_map[stream_url]
	if !ok {
		h = NewHub(
			stream_url,
			p.log,
		)

		p.streams_map[stream_url] = h
		go h.run()
	}
	h.register <- conn
	err := p.handler.OnConnect(conn)
	if err != nil {
		conn.error_channel <- err
	}

}

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
