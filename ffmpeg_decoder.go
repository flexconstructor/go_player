package go_player

import(


	. "github.com/3d0c/gmf"
	player_log "github.com/flexconstructor/go_player/log"
)


type FFmpegDecoder  struct{
	stream_url string
	broadcast chan []byte
	rtmp_status chan int
	codec_chan chan *CodecCtx
	error chan *WSError
	log player_log.Logger
	close_chan chan bool
	frame_channel chan *Frame
	packet_channel chan *Packet
}



func (d *FFmpegDecoder)Run(){
	d.log.Info("Run Decoder for %s",d.stream_url)
	defer close(d.frame_channel)
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

	for{
		select {
			case <- d.close_chan:
			d.log.Debug("close chan called")
			return
		default:
		packet:=inputCtx.GetNextPacket();
		if(packet != nil) {
			if packet.StreamIndex() == srcVideoStream.Index() {
				stream, err := inputCtx.GetStream(packet.StreamIndex())
				if (err != nil) {
					d.log.Error("can not decode stream")
					d.error <- NewError(13, 2)
				}else{
					for frame := range packet.Frames(stream.CodecCtx()) {
						d.log.Debug("frame")
						d.frame_channel <- frame.CloneNewFrame()
					}

					Release(packet)
				}
			}else{
				continue
			}

		}else{
			//d.log.Error("nil packet")
			continue
		}

	}

	}

	d.log.Info("Decoder stopped ")

}



func (d *FFmpegDecoder)Close(){
	d.log.Info("close decoder")
	d.close_chan <- true
	d.log.Debug("decoder closed")
//d.close_chan<-true
}
