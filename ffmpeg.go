package go_player

import (
	"github.com/3d0c/gmf"
	player_log "github.com/flexconstructor/go_player/log"
	"runtime"
	"sync"
	"fmt"
)

type ffmpeg struct {
	stream_url     string
	broadcast      chan []byte
	metadata       chan *MetaData
	error          chan *WSError
	log            player_log.Logger
	workers_length int
	close_channel  chan bool
	hub_id int
}
//Run ffmpeg functionality. Create decoder.
//Resave methadata. Run encoding to jpg when metadata income.
func (f *ffmpeg) run() {

	f.log.Info("run ffmpeg for %s", f.stream_url)
	runtime.GOMAXPROCS(runtime.NumCPU())
	codec_chan := make(chan *gmf.CodecCtx)
	frame_cannel := make(chan *gmf.Frame)
	decoder := &FFmpegDecoder{
		stream_url:     f.stream_url,
		codec_chan:     codec_chan,
		error:          f.error,
		log:            f.log,
		close_chan:     make(chan bool),
		frame_channel:  frame_cannel,
		packet_channel: make(chan *gmf.Packet),
		hub_id: f.hub_id,
	}
	go decoder.Run()
	defer decoder.Close()
	for {
		select {
		case <-f.close_channel:
		fmt.Println("close_chan income")
			return

		case c, ok := <-codec_chan:
		// set the metadata from codec.
			if ok {
				f.metadata <- &MetaData{
					Message: "metadata",
					Width:   c.Width(),
					Height:  c.Height(),
				}
				f.log.Debug("methadata: width= %d height= %d",c.Width(), c.Height())
				go f.runEncoder(c, frame_cannel)
			}
		}
	}
}
//Run encoding frames to jpeg images. Make pull of encoders.
func (f *ffmpeg) runEncoder(c *gmf.CodecCtx, frame_channel chan *gmf.Frame) {
	f.log.Debug("Run Encoder")
	wg := new(sync.WaitGroup)
	encoder := &FFmpegEncoder{
		srcCodec: c,
		broadcast: f.broadcast,
		error: f.error,
		log: f.log,
		close_chan: make(chan bool),
		frame_cannel: frame_channel,
		wg: wg,
		hub_id: f.hub_id,
	}

	for i := 0; i < f.workers_length; i++ {
		wg.Add(1)
		go encoder.Run()
	}
	f.log.Debug("all workers running")
	wg.Wait()
	f.log.Info("All encoders is done!")
}
//Close ffmpeg
func (f *ffmpeg) Close() {
	f.log.Info("Close ffmpeg for %s",f.stream_url)
	fmt.Println("close ffmpeg")
	f.close_channel <- true
}

func(f *ffmpeg) recoverFFMpeg(){
	if r := recover(); r != nil {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, false)
		reason := fmt.Sprintf("%v: %s", r, buf)
		f.log.Error("Runtime failure, reason -> %s", reason)
	}

}

