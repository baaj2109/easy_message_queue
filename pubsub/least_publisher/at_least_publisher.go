package least_publisher

import (
	"github.com/baaj2109/easy_message_queue/model"
	"github.com/baaj2109/easy_message_queue/pubsub"
)

type LeastUsedPublisher struct {
	SubscriberPriorityQueue
	pubsub.EmptyDone
}

func NewLeastUsedPublisher() *LeastUsedPublisher {
	publisher := LeastUsedPublisher{}
	publisher.items = make([]PriorityQueueNode, 1)
	publisher.subscribedClients = make(map[string]bool, 0)
	return &publisher
}

func (pq *LeastUsedPublisher) Publish(msg model.Message) bool {
	numSubscribers := pq.SubscriberCount()
	if numSubscribers > 0 {
		entry := pq.Pop()
		entry.Client.Notify(msg)
		entry.MessagesSent = entry.MessagesSent + 1
		pq.Push(entry)

		return true
	} else {
		return false
	}
}
