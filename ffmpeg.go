package go_player

import (
	"github.com/3d0c/gmf"
	player_log "github.com/flexconstructor/go_player/log"
	"runtime"
	"sync"
)

type ffmpeg struct {
	stream_url     string
	broadcast      chan []byte
	metadata       chan *MetaData
	error          chan *WSError
	log            player_log.Logger
	workers_length int
	close_channel  chan bool
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
	}
	defer decoder.Close()
	defer f.log.Debug("ffmpeg closed!!!")
	go decoder.Run()

	for {
		select {
		case <-f.close_channel:
			return

		case c, ok := <-codec_chan:
		// set the metadata from codec.
			if ok {
				f.metadata <- &MetaData{
					Message: "metadata",
					Width:   c.Width(),
					Height:  c.Height(),
				}
				go f.runEncoder(c, frame_cannel)
			}
		}
	}
}
//Run encoding frames to jpeg images. Make pull of encoders.
func (f *ffmpeg) runEncoder(c *gmf.CodecCtx, frame_channel chan *gmf.Frame) {
	wg := new(sync.WaitGroup)
	encoder := &FFmpegEncoder{
		srcCodec: c,
		broadcast: f.broadcast,
		error: f.error,
		log: f.log,
		close_chan: make(chan bool),
		frame_cannel: frame_channel,
		wg: wg,
	}

	for i := 0; i < f.workers_length; i++ {
		wg.Add(1)
		go encoder.Run()
	}
	wg.Wait()
	f.log.Info("All encoders is done!")
}
//Close ffmpeg
func (f *ffmpeg) Close() {
	f.log.Info("Close ffmpeg!")
	f.close_channel <- true
}
