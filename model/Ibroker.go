package model

type IBroker interface {
	Done()
	Publish(msg Message) bool
	Subscribe(client IClient) error
	Unsubscribe(client IClient)
	SubscriberCount() int
}
