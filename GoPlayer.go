package GoPlayer
import(
	"net/http"
	"github.com/gorilla/mux"
	"media-api/log"

)

var (
player_instance *GoPlayer=nil

)

type GoPlayer struct  {
	rtmp_url string
	streams_map map[string]*hub
	route *mux.Router
}

func NewGoPlayer() *GoPlayer{

	if player_instance==nil{
		player_instance=&GoPlayer{
			rtmp_url: "rtmp://"+GoPlayer_rtmp_host+":"+GoPlayer_rtmp_port+"/"+GoPlayer_app_name,
			streams_map: make(map[string]*hub),
			route: mux.NewRouter(),
		}
		http.Handle("/"+GoPlayer_app_name+"/",player_instance.route);
	}
	return player_instance
}

func(p *GoPlayer) Run(stream_name string) bool{
	if(p.streams_map[stream_name] == nil) {
		newhub:= NewHub(p.rtmp_url, stream_name)
		//p.route.HandleFunc("/"+GoPlayer_app_name+"/"+stream_name, serveWs)
		log.Debug("Run Go player: ",stream_name)
		p.streams_map[stream_name]=newhub
	//	go  newhub.run()
	}
	return true;
}

func (p *GoPlayer) Close(stream_name string)bool{
	h := p.streams_map[stream_name]
	if(h != nil){
		h.rtmp_status <-0
		return true
	}
return false
}

