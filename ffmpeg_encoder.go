package go_player
import (
	player_log "github.com/flexconstructor/go_player/log"
)

type FFmpegEncoder struct {
	metadata *MetaData
	broadcast chan []byte
	rtmp_status chan int
	error chan *WSError
	log player_log.Logger
	close_chan chan bool

}

func (e *FFmpegEncoder)Run(){
	e.log.Info("run encoder")
}