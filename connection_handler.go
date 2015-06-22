package go_player

type IConnectionHandler interface {
	OnConnect(conn *WSConnection)(*WSError)
	OnUpdate(conn *WSConnection)(*WSError)
	OnDisconnect(conn *WSConnection)(*WSError)
}