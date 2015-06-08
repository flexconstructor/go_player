package GoPlayer

import(
	rtmp "github.com/zhangpeihao/gortmp"
)

type RtmpHandler struct {
	stream_status chan int
	error_channel chan *Error
}
var status uint
var obConn rtmp.OutboundConn
var createStreamChan chan rtmp.OutboundStream
var videoDataSize int64
var audioDataSize int64


func (handler *RtmpHandler) OnStatus(conn rtmp.OutboundConn) {
	var err error
	status, err = obConn.Status()
	if(err != nil && handler.error_channel != nil){
		handler.error_channel <- NewError(8,2)
	}
}


func (handler *RtmpHandler) OnClosed(conn rtmp.Conn) {
	if(handler.error_channel != nil){
		handler.error_channel <- NewError(10,1)
	}

}

func (handler *RtmpHandler) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
	switch message.Type {
	case rtmp.VIDEO_TYPE:
		videoDataSize += int64(message.Buf.Len())
	case rtmp.AUDIO_TYPE:
		audioDataSize += int64(message.Buf.Len())
	}
}

func (handler *RtmpHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {

}

func (handler *RtmpHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	createStreamChan <- stream
	if(handler.stream_status != nil){
		handler.stream_status <- 1
	}
}


