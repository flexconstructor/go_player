package handlers
import "github.com/flexconstructor/go_player/ws"

type IConnectionHandler interface {
	OnConnect(conn *ws.WSConnection)(*ws.WSError)
	OnUpdate(conn *ws.WSConnection)(*ws.WSError)
	OnDisconnect(conn *ws.WSConnection)(*ws.WSError)
}