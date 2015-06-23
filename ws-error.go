package go_player
import "encoding/json"

type WSError struct {
	code uint8
	level uint8
	description string
}

var error_map = map[uint8]string {
	1: "No video stream found",
	2: "No codec found for this stream",
	3: "Can not open codec",
	4: "Can not decode stream",
	5: "Can not write decoded data",
	6: "The end of stream",
	7: "Can not play stream",
	8: "Can not check status of stream",
	9: "Can not connect to remote stream",
	10: "Connection closed from remote server",
	11: "Streaming server not running",
	12: "Can not decode packet",
	13: "Can not decode chunk",

}

func NewError(code uint8, level uint8)( *WSError){
	error_description, ok:=error_map[code]
	if(!ok){
		error_description="Unknown error"
	}
	er:= &WSError{
		code:code,
		level:level,
		description: error_description,
	}
	return er
}

func NewErrorWithDescription(code uint8, level uint8, description string)(*WSError){
	return &WSError{
		code:code,
		level:level,
		description: description,

}

}


func (er *WSError)JSON()([]byte, error){
	var m map[string]interface{}
	if(er.level==0) {
		m=map[string]interface{}{
			"success":1,
		}
	}else{
		m=map[string]interface{}{
			"error":map[string]interface{}{
				"code":er.code,
				"level":er.level,
				"description":er.description,
			},
		}
	}
	return json.Marshal(m)
}




