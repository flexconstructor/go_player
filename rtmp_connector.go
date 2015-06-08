package GoPlayer
import(
	rtmp "github.com/zhangpeihao/gortmp"
)



type RtmpConnector struct {
	rtmp_url string
	stream_id string
	handler *RtmpHandler
	error_cannel chan *Error

}

func (c *RtmpConnector)Run() {
	var err error
	createStreamChan = make(chan rtmp.OutboundStream)
	obConn, err = rtmp.Dial(c.rtmp_url, c.handler, 100)

	if err != nil && c.error_cannel != nil{
		c.error_cannel <- NewError(1,1)
		return ;
	}
	defer obConn.Close()

	err = obConn.Connect()
	if err != nil && c.error_cannel != nil{
		c.error_cannel <- NewError(1,1)
		return
	}
	for {
		select {
		case stream := <-createStreamChan:
		// Play
			err = stream.Play(c.stream_id, nil, nil, nil)
			if err != nil && c.error_cannel != nil{
				c.error_cannel <- NewError(7,1)
				return
			}
		}
	}



}

func (c *RtmpConnector)Close(){
	if(obConn != nil){
		obConn.Close()
	}
}