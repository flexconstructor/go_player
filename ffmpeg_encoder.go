package go_player
import (
	player_log "github.com/flexconstructor/go_player/log"
)

type FFmpegEncoder struct {
	broadcast chan []byte
	rtmp_status chan int
	error chan *WSError
	log player_log.Logger
	close_chan chan bool

}