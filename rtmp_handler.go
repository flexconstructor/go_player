package go_player

import(
	rtmp "github.com/zhangpeihao/gortmp"
	player_log "github.com/flexconstructor/go_player/log"
)

type RtmpHandler struct {
	stream_status chan int
	error_channel chan *WSError
	log player_log.Logger
}
var status uint
var obConn rtmp.OutboundConn
var createStreamChan chan rtmp.OutboundStream
var videoDataSize int64
var audioDataSize int64


func (handler *RtmpHandler) OnStatus(conn rtmp.OutboundConn) {
	var err error

	status, err = obConn.Status()
	if(err != nil && handler.error_channel != nil) {
		handler.error_channel <- NewError(8, 2)
		handler.log.Error("can not check status: %e",err)
	}else{
		handler.log.Info("rtmp status: %d",status)
	}

}


func (handler *RtmpHandler) OnClosed(conn rtmp.Conn) {
	handler.log.Info("stream closed")
	if(handler.error_channel != nil){
		handler.error_channel <- NewError(10,1)
	}

}

func (handler *RtmpHandler) OnReceived(conn rtmp.Conn, message *rtmp.Message) {

}

func (handler *RtmpHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {

}

func (handler *RtmpHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	createStreamChan <- stream
	handler.log.Info("On stream created: %d",handler.stream_status)
	if(handler.stream_status != nil){
		handler.stream_status <- 1
	}
}


