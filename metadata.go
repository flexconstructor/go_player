package go_player

import (
	"encoding/json"
)

type MetaData struct {
	Message string
	Width   int
	Height  int
}

func (m *MetaData) JSON() ([]byte, error) {
	return json.Marshal(m)
}
