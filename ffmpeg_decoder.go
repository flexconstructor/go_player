package go_player

import (
	. "github.com/3d0c/gmf"
	player_log "github.com/flexconstructor/go_player/log"
)
/*
Decode video stream from rtmp url to frames
 */
type FFmpegDecoder struct {
	stream_url     string             // rtmp stream url.
	codec_chan     chan *CodecCtx     // channel for codec.
	error          chan *WSError      // channel for error messages.
	log            player_log.Logger  // logger.
	close_chan     chan bool          // channel for close message.
	frame_channel  chan *Frame        // channel for frames
	packet_channel chan *Packet       // channel for packets of stream.
}

// Run decode
func (d *FFmpegDecoder) Run() {
	d.log.Info("Run Decoder for %s", d.stream_url)
	defer close(d.frame_channel)
	// create codec
	inputCtx, err := NewInputCtx(d.stream_url)

	if err != nil {
		d.error <- NewError(2, 1)
		return
	}
	defer inputCtx.CloseInputAndRelease()
	// get the video stream from flv container (without audio and methadata)
	srcVideoStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil {
		d.error <- NewError(1, 1)
		d.log.Error("stream not opend ")
		return
	}
	// send codec reference for execute of metadata.
	if srcVideoStream.CodecCtx() != nil {
		d.codec_chan <- srcVideoStream.CodecCtx()
	} else {
		d.log.Error("Invalid codec")
		d.error <- NewErrorWithDescription(1, 1, "Invalid codec")
		return
	}

	for {
		select {
		case <-d.close_chan:
			return
		default:
		// get next packet
			packet := inputCtx.GetNextPacket()
			if packet != nil {
				if packet.StreamIndex() == srcVideoStream.Index() {
					stream, err := inputCtx.GetStream(packet.StreamIndex())
					if err != nil {
						d.log.Error("can not decode stream")
						d.error <- NewError(13, 2)
					} else {
						// get next frame
						for frame := range packet.Frames(stream.CodecCtx()) {
							//d.frame_channel <- frame.CloneNewFrame()
							d.log.Debug("write new frame: %d",frame.TimeStamp())
							Release(frame)
						}

						Release(packet)
					}
				} else {
					continue
				}

			} else {
				continue
			}

		}

	}

}
// decoder close function.
func (d *FFmpegDecoder) Close() {
	d.log.Info("close decoder")
	d.close_chan <- true
}
