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
	metadata chan *MetaData
	error chan *WSError
	log player_log.Logger
	workers_length int
	close_channel chan bool

}




func (f *ffmpeg)run(){
	f.log.Info("run ffmpeg for %s",f.stream_url)
	runtime.GOMAXPROCS(runtime.NumCPU())
	codec_chan:= make(chan *gmf.CodecCtx)
	frame_cannel:= make(chan *gmf.Frame)
	decoder:=&FFmpegDecoder{
		stream_url: f.stream_url,
		broadcast:f.broadcast,
		codec_chan: codec_chan,
		error: f.error,
		log: f.log,
		close_chan: make(chan bool),
		frame_channel: frame_cannel,
		packet_channel: make(chan *gmf.Packet),
	}
	defer decoder.Close()
	defer f.log.Debug("ffmpeg closed!!!")
	go decoder.Run()

	for{
		select {
		case c, ok:= <-codec_chan:
		if(ok){
			f.metadata <- &MetaData{
				Message: "metadata",
				Width: c.Width(),
				Height: c.Height(),
			}
			f.runEncoder(c, frame_cannel)
		}

		case close, ok:= <-f.close_channel:
			f.log.Debug("call close_chan %t",close)
		if(ok ==true && close==true){
			f.log.Debug("CLOSE FFMPEG")
			return
		}

		/*case status := <- f.rtmp_status:
		f.log.Debug("call rtmp status: %d",status)
		if(status==0){
			f.log.Debug("rtmp status 0");
			return
		}
*/
		}
	}

}


func (f *ffmpeg)runEncoder(c *gmf.CodecCtx, frame_channel chan *gmf.Frame){
	f.log.Info("-----run encode-------r")
	wg:=new(sync.WaitGroup)
	encoder:=&FFmpegEncoder{
		c,
		f.broadcast,
		make(chan int),
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

func (f *ffmpeg)Close(){
	f.log.Info("Close ffmpeg!")
	f.close_channel <-true
	f.log.Debug("write status")

}



