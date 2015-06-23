package go_player

import(

	"runtime"
	. "github.com/3d0c/gmf"
	"runtime/debug"
	"sync"
	player_log "github.com/flexconstructor/go_player/log"
)

var broadcast chan []byte

func fatal(err error) {
	debug.PrintStack()


}


type FFmpegDecoder  struct{
	stream_url string
	broadcast chan []byte
	rtmp_status chan int
	metadata chan *MetaData
	error chan *WSError
	log player_log.Logger
}




func assert(i interface{}, err error) interface{} {
	if err != nil {
		fatal(err)
	}

	return i
}

func encodeWorker(data chan *Frame, wg *sync.WaitGroup, srcCtx *CodecCtx, error chan *WSError) {
	defer wg.Done()
	codec, err := FindEncoder("mjpeg")
	if err != nil && error != nil{
		fatal(err)
		error <- NewError(2,1)
		return
	}

	cc := NewCodecCtx(codec)
	defer Release(cc)

	w, h := srcCtx.Width(), srcCtx.Height()
	cc.SetPixFmt(AV_PIX_FMT_YUVJ420P).SetWidth(w).SetHeight(h)
	cc.SetWidth(w)
	cc.SetHeight(h)
	cc.SetTimeBase(srcCtx.TimeBase().AVR())

	if codec.IsExperimental() {
		cc.SetStrictCompliance(FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil && error != nil{
		fatal(err)
		error <- NewError(3,1)
		return
	}

	swsCtx := NewSwsCtx(srcCtx, cc, SWS_BICUBIC)
	defer Release(swsCtx)

	// convert to RGB, optionally resize could be here
	dstFrame := NewFrame().
	SetWidth(w).
	SetHeight(h).
	SetFormat(AV_PIX_FMT_YUVJ420P)
	defer Release(dstFrame)

	if err := dstFrame.ImgAlloc(); err != nil && error != nil{
		fatal(err)
		error <- NewError(4,2)
		return
	}

	for {
		srcFrame, ok := <-data
		if !ok {
			if(error != nil){
				error <- NewError(5,2)
			}
			Release(srcFrame)
			return
		}

		swsCtx.Scale(srcFrame, dstFrame)

		if p, ready, _ := dstFrame.EncodeNewPacket(cc); ready {
			if broadcast != nil {
				writeToBroadcast(p.Data());
			}else{
				return
			}

		}
		Release(srcFrame)
	}

}


func writeToBroadcast(b []byte){
	log.Debug("write bytes: %d",len(b))
	broadcast <-b

}


func (f *FFmpegDecoder)Run(){
	f.log.Info("RUN DECODER")
	broadcast=f.broadcast
	runtime.GOMAXPROCS(runtime.NumCPU())
	inputCtx := assert(NewInputCtx(f.stream_url)).(*FmtCtx)
	f.log.Info("open input codec: ",f.stream_url)
	defer inputCtx.CloseInputAndRelease()

	srcVideoStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil && f.error != nil{
		f.error <- NewError(1,1)
		f.log.Error("stream not opend ")
		return
	}
	f.log.Info("Open stream")
	if(f.metadata != nil && f.rtmp_status != nil){
		f.metadata <- &MetaData{
			Message: "metadata",
			Width: srcVideoStream.CodecCtx().Width(),
			Height: srcVideoStream.CodecCtx().Height(),
		}
		f.log.Info("write metadata")
	}

	wg := new(sync.WaitGroup)

	dataChan := make(chan *Frame)
	f.log.Info("run decoding")
	for i := 0; i < 20; i++ {
		wg.Add(i)
		go encodeWorker(dataChan, wg, srcVideoStream.CodecCtx(), f.error)
	}

	for packet := range inputCtx.GetNewPackets() {
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			f.log.Warn("Skip no video streams: %g",packet.StreamIndex())
			continue
		}

		ist := assert(inputCtx.GetStream(packet.StreamIndex())).(*Stream)

		for frame := range packet.Frames(ist.CodecCtx()) {
			dataChan <- frame.CloneNewFrame()
		}
		Release(packet)
	}

	if(f.error != nil){
		f.error <- NewError(6,1)
	}

	wg.Wait()

}

func (d *FFmpegDecoder)Close(){
d.log.Info("Close decoder")
}
