package main

import (
	"github.com/baaj2109/easy_message_queue/model"
	"github.com/baaj2109/easy_message_queue/pubsub/round_robin"
	"github.com/baaj2109/easy_message_queue/test"
)

func main() {

	client := test.NewTestClient("Subscription - 1")

	publisher := round_robin.NewRoundRobinPublisher()
	publisher.Subscribe(client)

	queue := model.NewGoQ(25, publisher)
	queue.StartPublishing()
	queue.Enqueue(model.Message{Id: "Message - 1"})
	queue.StopPublishing()
}
