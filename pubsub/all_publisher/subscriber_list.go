package all_publisher

import (
	"errors"
	"sync"

	"github.com/baaj2109/easy_message_queue/model"
)

type SubscriberList struct {
	lock  sync.Mutex
	Items []model.IClient
}

func NewSubscribersList() *SubscriberList {
	return &SubscriberList{
		Items: make([]model.IClient, 0),
	}
}

func (s *SubscriberList) Subscribe(client model.IClient) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.indexOf(client.Id()) >= 0 {
		return errors.New("duplicate client ids aren't allowed")
	}

	s.Items = append(s.Items, client)
	return nil
}

func (s *SubscriberList) Unsubscribe(client model.IClient) {
	s.lock.Lock()
	defer s.lock.Unlock()

	subIndex := s.indexOf(client.Id())
	if subIndex >= 0 {
		s.Items = append(s.Items[:subIndex], s.Items[subIndex+1:]...)
	}
}

func (s *SubscriberList) Get(index int) model.IClient {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Items[index]
}

func (s *SubscriberList) SubscriberCount() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.Items)
}

func (s *SubscriberList) indexOf(qClientId string) int {
	for i, item := range s.Items {
		if item.Id() == qClientId {
			return i
		}
	}
	return -1
}
