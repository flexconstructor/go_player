package GoPlayer
import(
	"net/http"
	"github.com/gorilla/mux"


	"strconv"
	"errors"
)

var (
player_instance *GoPlayer=nil
log Logger
)

// Represents logger with different levels of logs.
type Logger interface {
	Debug(interface{}, ...interface{})
	Trace(interface{}, ...interface{})
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{}) error
	Error(interface{}, ...interface{}) error
	Critical(interface{}, ...interface{}) error
	Close()
}

type GoPlayer struct  {
	rtmp_host string
	app_name string
	rtmp_port int
	http_port int
	streams_map map[string]*hub
	route *mux.Router
	log Logger
}


func InitGoPlayer(rtmp_host string, rtmp_port int, app_name string, http_port int, log Logger)*GoPlayer{

if(player_instance != nil){
	return player_instance
}
	player_instance=&GoPlayer{
		rtmp_host: rtmp_host,
		rtmp_port: rtmp_port,
		app_name:app_name,
		http_port: http_port,
		streams_map: make(map[string]*hub),
		route: mux.NewRouter(),
		log: log,
	}
	player_instance.route.Headers("Access-Control-Allow-Origin","*")
	http.Handle("/"+player_instance.app_name+"/",player_instance.route);
	http.ListenAndServe(":"+strconv.Itoa(player_instance.http_port),nil);
	log.Debug("Init player instance rtmp hodt: ",player_instance.rtmp_host)
	log.Debug("Init player instance rtmp port: ",player_instance.rtmp_port)
	log.Debug("Init player instance rtmp app: ",player_instance.app_name)
	log.Debug("Init player instance http port: ",player_instance.http_port)
	log.Debug("---------------------")
	return player_instance
}

func GetPlayerInstance()( *GoPlayer,error){
	if(player_instance==nil){
		return nil,errors.New("GoPlayer not initialized!")
	}
	return player_instance, nil
}


func CloseGoPlayer()(error, bool){

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
		}
	}
	player.log.Close()
	return nil, true
}



/*func NewGoPlayer() *GoPlayer{

	if player_instance==nil{
		player_instance=&GoPlayer{
			rtmp_url: "rtmp://"+GoPlayer_rtmp_host+":"+GoPlayer_rtmp_port+"/"+GoPlayer_app_name,
			streams_map: make(map[string]*hub),
			route: mux.NewRouter(),
		}
		player_instance.route.Headers("Access-Control-Allow-Origin","*")
		http.Handle("/"+GoPlayer_app_name+"/",player_instance.route);
		http.ListenAndServe(":"+strconv.Itoa(GoPlayer_ws_port),nil)
	}
	return player_instance
}*/

func(p *GoPlayer) Run(stream_name string) bool{

	p.log.Info("Run player with stream: ",stream_name)
	p.log.Info("APPName: ",p.app_name)
	if(p.streams_map[stream_name] == nil) {
		//newhub:= NewHub("rtmp://"+p.rtmp_host+":"+strconv.Itoa(p.rtmp_port)+"/"+p.app_name, stream_name,p.log)
		p.route.HandleFunc("/"+p.app_name+"/"+stream_name, serveWs)
		//p.streams_map[stream_name]=newhub
		//go  newhub.run()
	}
	return true;
}

func (p *GoPlayer) Close(stream_name string)bool{
	log.Info("CLOSE STREAM: ",stream_name)
	defer log.Close()
	h := p.streams_map[stream_name]
	if(h != nil){
		h.rtmp_status <-0
		return true
	}

return false
}

