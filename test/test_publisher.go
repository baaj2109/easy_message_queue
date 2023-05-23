package test

import (
	"github.com/baaj2109/easy_message_queue/model"
	"github.com/baaj2109/easy_message_queue/pubsub/all_publisher"
)

type TestPublisher struct {
	Responses   chan bool
	Messages    chan model.Message
	DoneCalls   chan bool
	Subscribers chan *all_publisher.SubscriberList
}

func NewTestPublisher() *TestPublisher {
	return &TestPublisher{
		Responses: make(chan bool, 1000),
		Messages:  make(chan model.Message, 1000),
		DoneCalls: make(chan bool, 1000),
	}
}

func (tp *TestPublisher) Publish(msg model.Message) bool {
	tp.Messages <- msg

	if len(tp.Responses) == 0 {
		panic("responses must be set on the test publisher")
	}

	return <-tp.Responses
}

func (tp *TestPublisher) Done() {
	tp.DoneCalls <- true
}

func (tp *TestPublisher) Subscribe(client model.IClient) error {
	return nil
}

func (tp *TestPublisher) Unsubscribe(client model.IClient) {}

func (tp *TestPublisher) SubscriberCount() int {
	return 0
}
