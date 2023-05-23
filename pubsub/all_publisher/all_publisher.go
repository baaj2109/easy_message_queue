package all_publisher

import (
	"github.com/baaj2109/easy_message_queue/model"
	"github.com/baaj2109/easy_message_queue/pubsub"
)

type AllPublisher struct {
	SubscriberList
	pubsub.EmptyDone
}

func NewAllPublisher() *AllPublisher {
	publisher := AllPublisher{}
	publisher.Items = make([]model.IClient, 0)

	return &publisher
}

func (ap *AllPublisher) Publish(msg model.Message) bool {
	numSubScribers := ap.SubscriberCount()
	delivered := false

	for i := 0; i < numSubScribers; i++ {
		ap.Get(i).Notify(msg)
		delivered = true
	}

	return delivered
}
