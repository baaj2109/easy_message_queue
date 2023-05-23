package round_robin

import (
	"github.com/baaj2109/easy_message_queue/model"
	"github.com/baaj2109/easy_message_queue/pubsub"
	"github.com/baaj2109/easy_message_queue/pubsub/all_publisher"
)

type RoundRobinPublisher struct {
	all_publisher.SubscriberList
	pubsub.EmptyDone
	nextNotified int
}

func NewRoundRobinPublisher() *RoundRobinPublisher {
	publisher := RoundRobinPublisher{
		nextNotified: -1,
	}
	publisher.Items = make([]model.IClient, 0)
	return &publisher
}

func (rr *RoundRobinPublisher) Publish(msg model.Message) bool {
	numSubScribers := rr.SubscriberCount()
	if numSubScribers > 0 {
		rr.nextNotified = (rr.nextNotified + 1) % numSubScribers
		rr.Get(rr.nextNotified).Notify(msg)
	} else {
		return false
	}

	return true
}
