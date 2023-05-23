package model

import (
	"errors"
	"sync"
)

type GoQ struct {
	broker    IBroker
	maxDepth  int
	queue     chan Message
	pauseChan chan bool
	once      sync.Once
	done      bool
}

func NewGoQ(queueDepth int, publisher IBroker) *GoQ {
	return &GoQ{
		maxDepth:  queueDepth,
		queue:     make(chan Message, queueDepth),
		pauseChan: make(chan bool, 1),
		broker:    publisher,
	}
}

func (q *GoQ) Enqueue(message Message) error {
	if q.done {
		return errors.New("Queue closed")
	}

	select {
	case q.queue <- message:
		return nil
	default:
		return errors.New("Message rejected, max queue depth reached")
	}
}

func (q *GoQ) StartPublishing() {
	go func() {
		for {
			msg, ok := <-q.queue
			if ok {
				select {
				case <-q.pauseChan:
					return
				default:
					q.publishMessage(msg)
				}
			} else {
				q.broker.Done()
				return
			}
		}
	}()
}

func (q *GoQ) StopPublishing() {
	q.once.Do(func() {
		close(q.queue)
		q.done = true
	})
}

func (q *GoQ) publishMessage(msg Message) {
	delivered := q.broker.Publish(msg)
	if !delivered {
		q.queue <- msg
	}
}

func (q *GoQ) PausePublishing() {
	q.pauseChan <- true
}
