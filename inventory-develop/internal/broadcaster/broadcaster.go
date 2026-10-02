package broadcaster

type Broadcaster interface {
	Broadcast(message []byte)
}
