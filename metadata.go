package go_player

import (
	"encoding/json"
)
// Metadata object for setting width/height for stream client.
type MetaData struct {
	Message string
	Width   int
	Height  int
}

func (m *MetaData) JSON() ([]byte, error) {
	return json.Marshal(m)
}
