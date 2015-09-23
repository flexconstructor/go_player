package go_player

//Interface for represent event handling of go player
type IConnectionHandler interface {

	//On new client is connected
	OnConnect(conn *WSConnection) *WSError

	// On stream update
	OnUpdate(conn *WSConnection) *WSError

	//On client is disconnected
	OnDisconnect(conn *WSConnection) *WSError
}
