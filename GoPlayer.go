package GoPlayer
import(
	"net/http"
	"github.com/gorilla/mux"

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
			rtmp_url: "rtmp://"+rtmp_host+":"+rtmp_port+"/"+app_name,
			streams_map: make(map[string]*hub),
			route: mux.NewRouter(),
		}
		http.Handle("/"+app_name+"/",player_instance.route);
	}
	return player_instance
}

func(p *GoPlayer) Run(stream_name string) bool{
	if(p.streams_map[stream_name] == nil) {
		newhub:= NewHub(p.rtmp_url, stream_name)
		p.route.HandleFunc("/"+app_name+"/"+stream_name, serveWs)
		p.streams_map[stream_name]=newhub
		go  newhub.run()
	}

	return true;
}
