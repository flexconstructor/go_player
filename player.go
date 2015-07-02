package go_player

import(

	"errors"
	player_log "github.com/flexconstructor/go_player/log"
	"strconv"
)

var (
player_instance *GoPlayer=nil
log player_log.Logger
)




type GoPlayer struct  {
	rtmp_host string
	app_name string
	rtmp_port int
	http_port int
	streams_map map[uint64]*hub
	streams chan *hub
	connects chan *WSConnection
	updates chan  *WSConnection
	closes chan *WSConnection
	stops chan *GoPlayer
	hub_close chan *hub
	log player_log.Logger
	service_token string

	handler IConnectionHandler
}


func InitGoPlayer(
rtmp_host string,
rtmp_port int,
app_name string,
http_port int,
log player_log.Logger,
service_token string,
connectionHandler IConnectionHandler)*GoPlayer{

if(player_instance != nil){
	return player_instance
}
	player_instance=&GoPlayer{
		rtmp_host: rtmp_host,
		rtmp_port: rtmp_port,
		app_name:app_name,
		http_port: http_port,
		log: log,
		streams_map: make(map[uint64]*hub),
		streams: make(chan *hub),
		connects:make(chan *WSConnection),
		updates:make(chan *WSConnection),
		closes: make(chan *WSConnection),
		stops: make(chan *GoPlayer),
		hub_close: make(chan *hub),
		service_token: service_token,
		handler: connectionHandler,
	}

	log.Debug("Init player instance rtmp host: %s",player_instance.rtmp_host)
	log.Debug("Init player instance rtmp port: %d",player_instance.rtmp_port)
	log.Debug("Init player instance rtmp app: %s",player_instance.app_name)
	log.Debug("Init player instance http port: %d",player_instance.http_port)
	log.Debug("---------------------")
	return player_instance
}

func GetPlayerInstance()( *GoPlayer,error){
	if(player_instance==nil){
		return nil,errors.New("goplayer not initialized!")
	}
	return player_instance, nil
}


func (p *GoPlayer)Run(){
p.log.Info("Run GO PLAYER INSTANCE")
defer p.stopInstance()
	for {
		select {
		case <-p.stops:
			return
		case c,ok:=<- p.connects:
		p.log.Debug("call write connection")
			if(ok) {
				p.initConnection(c)
			}else{
				p.log.Error("can not write connection")
			}
		case c,ok:= <- p.closes:
		if(ok) {
			p.closeConnection(c)
		}else{
			p.log.Error("can not close connection")
		}

		case h,ok:= <-p.hub_close:
		if(ok){
			p.log.Debug("remove hub from map")
			streamID,err:=strconv.ParseUint(h.stream_id, 10, 64)
			if(err != nil){
				panic(err)
			}
			delete(p.streams_map, streamID)
		}
		case u,ok:= <- p.updates:
		if(!ok){
			panic("can not write update")
		}
		err:=p.handler.OnUpdate(u)
		if(err != nil){
			u.error_channel <- err
		}

		}

		}
}

func (p *GoPlayer)Stop(){
p.log.Info("STOP GO PLAYER INSTANCE")
	p.stops<-p


}


func (p *GoPlayer)stopInstance(){
	p.log.Info("Player stopped")
	player_instance=nil;

}


func (p *GoPlayer)initConnection(conn *WSConnection){
	params:=conn.GetConnectionParameters()
	p.log.Debug("init connection  with params: stream_id=  %d user_id= %d access_token= %s",params.StreamID, params.ClientID, params.AccessToken)
	h,ok:=p.streams_map[params.StreamID]
	if(!ok){
		p.log.Debug("init hub")
		h=NewHub("rtmp://"+p.rtmp_host+":"+strconv.Itoa(p.rtmp_port)+"/"+p.app_name,
			strconv.FormatUint(params.StreamID,10),
			p.log,
			p.service_token,
			p.hub_close,
		)
		p.streams_map[params.StreamID]=h
		go h.run()
	}
	p.log.Debug("register connection in hub")
	h.register<-conn
	err:= p.handler.OnConnect(conn)
	if(err != nil){
		conn.error_channel <-err
	}

}

func (p *GoPlayer)closeConnection(conn *WSConnection){
	params:=conn.GetConnectionParameters()
	p.log.Debug("Close connection with params: stream_id=  %d user_id= %d access_token= %s",params.StreamID, params.ClientID, params.AccessToken)
	h,ok:=p.streams_map[params.StreamID]
	if(! ok){
		p.log.Error("hub for stream %d not found!",params.StreamID)
		return
	}
	h.unregister <- conn
	err:=p.handler.OnDisconnect(conn)
	if(err != nil){
		conn.error_channel <- err
	}
	p.log.Debug("live connections: %d",len(h.connections))
	if(len(h.connections)==0){
		delete(p.streams_map, params.StreamID)
		p.log.Debug("remove hub from map",params.StreamID)
	}
}


