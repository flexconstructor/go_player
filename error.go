package  go_player
import "encoding/json"

type Error struct {
	Message string
	Code int
	Level int
	Description string
}

var error_map = map[int]string {
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

}

func (er *Error)JSON()([]byte, error){
	return json.Marshal(er)
}

func NewError(code int, level int)( *Error){
	error_description, ok:=error_map[code]
	if(!ok){
		error_description="Unknown error"
	}
	er:= &Error{
		Message:"error",
		Code:code,
		Level:level,
		Description: error_description,
	}
return er
}