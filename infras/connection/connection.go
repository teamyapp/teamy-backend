package connection

type Connection interface {
	OnErrors() <-chan error
	OnMessageReceived() <-chan []byte
	SendMessage(message []byte)
	OnClientDisconnect() <-chan bool
	Close() error
}
