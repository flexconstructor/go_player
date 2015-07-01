package go_player
import (
	player_log "github.com/flexconstructor/go_player/log"
	"runtime"
	"github.com/3d0c/gmf"
	"sync"
)
type ffmpeg struct {
	stream_url string
	broadcast chan []byte
	rtmp_status chan int
	metadata chan *MetaData
	error chan *WSError
	log player_log.Logger
	workers_length int

}




func (f *ffmpeg)run(){
	f.log.Info("run ffmpeg for %s",f.stream_url)
	runtime.GOMAXPROCS(runtime.NumCPU())
	codec_chan:= make(chan *gmf.CodecCtx)
	frame_cannel:= make(chan *gmf.Frame)
	decoder:=&FFmpegDecoder{
		f.stream_url,
		f.broadcast,
		f.rtmp_status,
		codec_chan,
		f.error,
		f.log,
		make(chan bool),
		frame_cannel,
		make(chan *gmf.Packet),
	}
	defer decoder.Close()
	go decoder.Run()

	for{
		select {
		case c, ok:= <-codec_chan:
		if(ok){
			//f.log.Info("ON METADATA w= %w h= %h",m.Width, m.Height )

			f.metadata <- &MetaData{
				Message: "metadata",
				Width: c.Width(),
				Height: c.Height(),
			}
			f.runEncoder(c, frame_cannel)
		}

		case status := <- f.rtmp_status:
		f.log.Debug("call rtmp status: %d",status)
		if(status==0){
			f.log.Debug("rtmp status 0");
			return
		}

		}
	}

}


func (f *ffmpeg)runEncoder(c *gmf.CodecCtx, frame_channel chan *gmf.Frame){
	f.log.Info("-----run encode-------r")
	wg:=new(sync.WaitGroup)
	encoder:=&FFmpegEncoder{
		c,
		f.broadcast,
		f.rtmp_status,
		f.error,
		f.log,
		make(chan bool),
		frame_channel,
		wg,
	}

	for i:=0;i<f.workers_length ;i++  {
		wg.Add(1)
		go encoder.Run()
	}

	wg.Wait()
	f.log.Info("All encoders is done!")


}

func (f *ffmpeg)close(){
	f.log.Info("Close ffmpeg!")
	f.rtmp_status <-0

}



