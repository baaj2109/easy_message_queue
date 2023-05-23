package test

import "github.com/baaj2109/easy_message_queue/model"

type TestClient struct {
	ClientId      string
	Notifications chan model.Message
	Errors        chan error
}

func NewTestClient(id string) *TestClient {
	return &TestClient{
		ClientId:      id,
		Notifications: make(chan model.Message, 1000),
	}
}

func (qc *TestClient) Id() string {
	return qc.ClientId
}

func (qc *TestClient) Notify(message model.Message) error {
	qc.Notifications <- message
	return nil
}
