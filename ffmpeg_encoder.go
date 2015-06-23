package go_player
import (
	player_log "github.com/flexconstructor/go_player/log"
	"github.com/3d0c/gmf"
)

type FFmpegEncoder struct {
	srcCodec *gmf.CodecCtx
	broadcast chan []byte
	rtmp_status chan int
	error chan *WSError
	log player_log.Logger
	close_chan chan bool
	frame_cannel chan *gmf.Frame

}

func (e *FFmpegEncoder)Run(){
	e.log.Info("run encoder")
	codec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_MJPEG )
	if(err != nil){
		e.error <- NewError(2,1)
		return
	}

	cc := gmf.NewCodecCtx(codec)
	defer gmf.Release(cc)
	cc.SetPixFmt(gmf.AV_PIX_FMT_YUVJ420P)
	cc.SetWidth(e.srcCodec.Width())
	cc.SetHeight(e.srcCodec.Height())
	cc.SetTimeBase(e.srcCodec.TimeBase().AVR())

	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		e.log.Error("can not open codec")
		e.error <- NewError(3,1)
		return
	}

	swsCtx := gmf.NewSwsCtx(e.srcCodec, cc, gmf.SWS_BICUBIC)
	defer gmf.Release(swsCtx)

	// convert to RGB, optionally resize could be here
	dstFrame := gmf.NewFrame().
	SetWidth(e.srcCodec.Width()).
	SetHeight(e.srcCodec.Height()).
	SetFormat(gmf.AV_PIX_FMT_YUVJ420P)
	e.log.Debug("setting format")
	defer gmf.Release(dstFrame)

	if err := dstFrame.ImgAlloc(); err != nil {
		e.log.Error("codec error: ",err)
		e.error <- NewError(4,2)
		return
	}

	e.log.Debug("wait data from decoder...")
	for {
		srcFrame, ok := <-e.frame_cannel
		if !ok {

			e.log.Error("frame error: ")
			e.error <- NewError(5,2)
			gmf.Release(srcFrame)
			return
		}

		swsCtx.Scale(srcFrame, dstFrame)
		e.log.Debug("scale frame")
		if p, ready, _ := dstFrame.EncodeNewPacket(cc); ready {
			e.log.Debug("frame ready")
			e.broadcast <-p.Data()
		}
		e.log.Debug("release frame after")
		gmf.Release(srcFrame)
	}

}