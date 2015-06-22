package go_player

import(

	"errors"
	"github.com/flexconstructor/go_player/ws"
	player_log "github.com/flexconstructor/go_player/log"
	handler "github.com/flexconstructor/go_player/handlers"
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
	connects chan *ws.WSConnection
	updates chan  *ws.WSConnection
	closes chan *ws.WSConnection
	stops chan *GoPlayer
	log player_log.Logger
	service_token string
	handler handler.IConnectionHandler
}


func InitGoPlayer(
rtmp_host string,
rtmp_port int,
app_name string,
http_port int,
log player_log.Logger,
service_token string,
connectionHandler handler.IConnectionHandler)*GoPlayer{

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
		connects:make(chan *ws.WSConnection),
		updates:make(chan *ws.WSConnection),
		closes: make(chan *ws.WSConnection),
		stops: make(chan *GoPlayer),
		service_token: service_token,
		handler: connectionHandler,
	}

	log.Debug("Init player instance rtmp hodt: ",player_instance.rtmp_host)
	log.Debug("Init player instance rtmp port: ",player_instance.rtmp_port)
	log.Debug("Init player instance rtmp app: ",player_instance.app_name)
	log.Debug("Init player instance http port: ",player_instance.http_port)
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
			if(ok){
				p.initConnection(c)
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

func (p *GoPlayer)RegisterConnection(conn *ws.WSConnection){
	p.log.Debug("register connection: ")
	p.connects <-conn

}

func (p *GoPlayer)initConnection(conn *ws.WSConnection){
	params:=conn.GetConnectionParameters()
	h,ok:=p.streams_map[params.StreamID]
	if(!ok){
		h=NewHub("rtmp://"+p.rtmp_host+":"+strconv.Itoa(p.rtmp_port)+"/"+p.app_name,
			strconv.FormatUint(params.StreamID,10),
			p.log,
			p.service_token,
		)
		go h.run()
	}
	h.register<-conn
	p.handler.OnConnect(conn)

}


/*func CloseGoPlayer()(error, bool){

	player, err:= GetPlayerInstance()
	if(err != nil){
		return err,false;
	}
	player.log.Info("Close Go Player")

	for stream_name := range player.streams_map {
			h:=player.streams_map[stream_name]
			if(h != nil){
				h.rtmp_status <-0;
				delete(player.streams_map,stream_name)
				h=nil
		}
	}
	player_instance=nil
	player.log.Close()
	return nil, true
}*/


/*func(p *GoPlayer) Run(stream_name string) bool{

	p.log.Info("Run player with stream: ",stream_name)
	p.log.Info("APPName: ",p.app_name)
	if(p.streams_map[stream_name] == nil) {
		p.log.Debug("handle ws function: "+"/"+p.app_name+"/"+stream_name)
		//newhub:= NewHub("rtmp://"+p.rtmp_host+":"+strconv.Itoa(p.rtmp_port)+"/"+p.app_name, stream_name,p.log,p.service_token,p.handler)
		p.route.HandleFunc("/"+stream_name, p.serveWebSocket)
		//p.route.NewRoute().HandlerFunc("/"+stream_name, p.serveWebSocket)
		p.log.Debug("-----")
	//	p.streams_map[stream_name]=newhub
		//go  newhub.run()
	}
	return true;
}*/

/*func (p *GoPlayer)GetStream(stream_id uint64)(*hub){

return nil
}

func (p *GoPlayer) Close(stream_id uint64)bool{
	/*p.log.Debug("CLOSE STREAM: ",stream_name)
	p.route.HandleFunc("/"+p.app_name+"/"+stream_name, nil)
	h := p.streams_map[stream_name]
	if(h != nil){
		h.rtmp_status <-0
		return true
	}
*/
//return false
//}

