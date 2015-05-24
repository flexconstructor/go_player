package GoPlayer
import(
	"log"
)

type GoPlayer struct  {
	rtmp_url string

}

func NewGoPlayer(rtmp_url string) *GoPlayer{
	return &GoPlayer{rtmp_url}
}

func(p *GoPlayer) Run(stream_name string) bool{
	log.Print("run go player>>>: ",p.rtmp_url+"/"+stream_name)
	return true;
}

