package  go_player
import(
	rtmp "github.com/zhangpeihao/gortmp"
	player_log "github.com/flexconstructor/go_player/log"

)



type RtmpConnector struct {
	rtmp_url string
	stream_id string
	handler *RtmpHandler
	error_cannel chan *WSError
	close_channel chan bool
	log player_log.Logger

}

func (c *RtmpConnector)Run() {
	var err error
	c.log.Info("Run RTMP connection: ",c.rtmp_url)
	createStreamChan = make(chan rtmp.OutboundStream)
	obConn, err = rtmp.Dial(c.rtmp_url, c.handler, 100)

	if err != nil && c.error_cannel != nil{
		c.log.Error("no stream error: ",err)
		c.error_cannel <- NewError(1,1)
		return ;
	}


	err = obConn.Connect()
	if err != nil && c.error_cannel != nil{
		c.log.Error("connection error: ",err)
		c.error_cannel <- NewError(1,1)
		return
	}
	defer obConn.Close()
	for {
		select {
		case stream := <-createStreamChan:
		// Play
			c.log.Info("Play stream: ",c.stream_id)
			err = stream.Play(c.stream_id, nil, nil, nil)
			if err != nil && c.error_cannel != nil{
				c.log.Error("Can not play this stream")
				c.error_cannel <- NewError(7,1)
				return
			}
		case b:=<-c.close_channel:
		if(b==true){
			c.log.Debug("calling close channel")
			return
		}
		}
	}



}

func (c *RtmpConnector)Close(){
	c.log.Info("close connection")
	/*if(obConn != nil){
		obConn.Close()
	}*/
	c.close_channel <-true
}