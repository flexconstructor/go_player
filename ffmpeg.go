package go_player
import (
	player_log "github.com/flexconstructor/go_player/log"
	"runtime"
)
type ffmpeg struct {
	stream_url string
	broadcast chan []byte
	rtmp_status chan int
	metadata chan *MetaData
	error chan *WSError
	log player_log.Logger
	close_chan chan bool
	workers_length int

}




func (f *ffmpeg)run(){
	f.log.Info("run ffmpeg for %s",f.stream_url)
	defer f.close()
	runtime.GOMAXPROCS(runtime.NumCPU())
	md:= make(chan *MetaData)
	decoder:=&FFmpegDecoder{
		f.stream_url,
		f.broadcast,
		f.rtmp_status,
		md,
		f.error,
		f.log,
		f.close_chan,
	}

	go decoder.Run()

	for{
		select {
		case m, ok:= <- md:
		if(ok){
			f.log.Info("ON METADATA w= %w h= %h",m.Width, m.Height )
			f.runEncoder(m)
			f.metadata <- m
		}
		case _, ok:= <- f.close_chan:
		if(ok){
			return
		}
		}
	}



}


func (f *ffmpeg)runEncoder(m *MetaData){
	f.log.Info("run encoder")
	encoder:=&FFmpegEncoder{
		m,
		f.broadcast,
		f.rtmp_status,
		f.error,
		f.log,
		f.close_chan,
	}

	for i:=0;i<f.workers_length ;i++  {
		go encoder.Run()
	}

}

func (f *ffmpeg)close(){
	f.log.Info("Close ffmpeg!")

}



