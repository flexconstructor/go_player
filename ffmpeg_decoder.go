package go_player

import(


	. "github.com/3d0c/gmf"
	player_log "github.com/flexconstructor/go_player/log"
)

/*var broadcast chan []byte

func fatal(err error) {
	debug.PrintStack()
	log.Error("fatal")

}*/


type FFmpegDecoder  struct{
	stream_url string
	broadcast chan []byte
	rtmp_status chan int
	codec_chan chan *CodecCtx
	error chan *WSError
	log player_log.Logger
	close_chan chan bool
	frame_channel chan *Frame
}



func (d *FFmpegDecoder)Run(){
	d.log.Info("Run Decoder for %s",d.stream_url)
	inputCtx,err:=NewInputCtx(d.stream_url)

	if(err != nil){
		d.error <-NewError(2,1)
		return
	}
	defer inputCtx.CloseInputAndRelease()
	srcVideoStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil{
		d.error <- NewError(1,1)
		d.log.Error("stream not opend ")
		return
	}


	if(srcVideoStream.CodecCtx() != nil) {
		d.codec_chan <- srcVideoStream.CodecCtx()
	}else{
		d.log.Error("Invalid codec")
		d.error<-NewErrorWithDescription(1,1,"Invalid codec")
		return
	}
	packets:= inputCtx.GetNewPackets()
	for packet := range  packets{
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			d.log.Warn("Skip no video streams: %g",packet.StreamIndex())
			continue
		}

			stream, err := inputCtx.GetStream(packet.StreamIndex())

			if (err != nil) {
				d.log.Error("can not decode stream")
				d.error <- NewError(13, 1)

			}
		d.log.Debug("stream: is video: %b Duration: %d",stream.IsVideo(),srcVideoStream.Duration())
			for frame := range packet.Frames(stream.CodecCtx()) {
				d.log.Info("new frame: ",frame.TimeStamp())
				d.frame_channel <- frame.CloneNewFrame()

			}
		Release(packet)

	}
	
	d.log.Info("Decoder stopped index %d",srcVideoStream.Index())

}

func (d *FFmpegDecoder)Close(){
	d.log.Info("close decoder")
d.close_chan<-true
}
/*
func encodeWorker(data chan *Frame, wg *sync.WaitGroup, srcCtx *CodecCtx, error chan *WSError, logger player_log.Logger) {
	defer wg.Done()
	logger.Debug("run worker")
	codec, err := FindEncoder(AV_CODEC_ID_MJPEG )
	if err != nil{
		logger.Error("can not find codec %e",err)
		error <- NewError(2,1)
		return
	}
	logger.Debug("codec find: ",codec)
	cc := NewCodecCtx(codec)
	defer Release(cc)
	logger.Debug("new codec ctx: ",cc)
	w, h := srcCtx.Width(), srcCtx.Height()
	cc.SetPixFmt(AV_PIX_FMT_YUVJ420P).SetWidth(w).SetHeight(h)
	cc.SetWidth(w)
	cc.SetHeight(h)
	cc.SetTimeBase(srcCtx.TimeBase().AVR())
	logger.Debug("codec settings created")
	if codec.IsExperimental() {
		cc.SetStrictCompliance(FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		logger.Error("can not open codec")
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
	logger.Debug("setting format")
	defer Release(dstFrame)

	if err := dstFrame.ImgAlloc(); err != nil {
		logger.Error("codec error: ",err)
		error <- NewError(4,2)
		return
	}
		logger.Debug("wait data from decoder...")
	for {
		srcFrame, ok := <-data
		if !ok {

				logger.Error("frame error: ")
				error <- NewError(5,2)

			logger.Debug("release frame")
			Release(srcFrame)
			return
		}

		swsCtx.Scale(srcFrame, dstFrame)
		logger.Debug("scale frame")
		if p, ready, _ := dstFrame.EncodeNewPacket(cc); ready {
			logger.Debug("frame ready")
				writeToBroadcast(p.Data());
		}
		logger.Debug("release frame after")
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
	/*inputCtx := assert(NewInputCtx(f.stream_url)).(*FmtCtx)
	f.log.Info("open input codec: ",f.stream_url)
	inputCtx,err:=NewInputCtx(f.stream_url)
	if(err != nil){
		f.log.Error("can not find codec")
		f.error <- NewError(2,1)
		return
	}
	packet_chan:=inputCtx.GetNewPackets()

	for {
	 select {
		case packet,ok:=<-packet_chan:
		if(ok){
			f.log.Debug("packet: ",packet)
		}
		}
	}
	defer inputCtx.CloseInputAndRelease()

	srcVideoStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil && f.error != nil{
		f.error <- NewError(1,1)
		f.log.Error("stream not opend ")
		return
	}
	f.log.Info("Open stream")
		if(srcVideoStream.CodecCtx() != nil) {
			f.metadata <- &MetaData{
				Message: "metadata",
				Width: srcVideoStream.CodecCtx().Width(),
				Height: srcVideoStream.CodecCtx().Height(),
			}
			f.log.Info("write metadata")
		}else{
			f.log.Error("Invalid codec")
		f.error<-NewErrorWithDescription(1,1,"Invalid codec")
			return
		}


	/*for packet := range inputCtx.GetNewPackets() {
		log.Debug("new Packet ",packet.StreamIndex())
	}
	dataChan := make(chan *Frame)
	wg := new(sync.WaitGroup)
	for i := 0; i < 20; i++ {
		f.log.Debug("run worker",i)
		wg.Add(i)
		go encodeWorker(dataChan, wg, srcVideoStream.CodecCtx(), f.error, f.log)

	}
	wg.Wait()
	f.log.Debug("all workers close: ")
	/*for packet := range inputCtx.GetNewPackets() {
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
		f.log.Error("write error no stream")
		return
		//f.error <- NewError(6,1)
	}



}

func (d *FFmpegDecoder)Close(){
d.log.Info("Close decoder")
}
*/