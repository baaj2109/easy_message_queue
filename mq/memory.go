package mq

import (
	"sync"
	"time"
)

// MemoryMessageQueue 内存实现的处理池本体
type MemoryMessageQueue struct {
	topics    map[string][]chan Message // topics 用于存放所有的 channel 订阅者
	callbacks map[string][]CallbackFunc // callbacks 用于存放所有的 callback function 订阅者
	sync.RWMutex
}

var _ Pool = &MemoryMessageQueue{}

func NewMemoryMQ() Pool {
	return &MemoryMessageQueue{
		topics:    make(map[string][]chan Message),
		callbacks: make(map[string][]CallbackFunc),
	}
}

// Publish 发布消息
func (m *MemoryMessageQueue) Publish(topic string, message Message) {
	// 锁定全局队列，获取所有的订阅者
	m.RLock()
	subscribersChan, okChan := m.topics[topic]
	subscribersCallback, okCallback := m.callbacks[topic]
	m.RUnlock()
	// 这是为了避免在发布消息的时候对订阅者进行修改，从而导致订阅者的遍历出现问题

	if okChan {
		go func(subscribersChan []chan Message) {
			for i := 0; i < len(subscribersChan); i++ {
				select {
				case subscribersChan[i] <- message:
				case <-time.After(time.Second):
				}
			}
		}(subscribersChan)

	}

	if okCallback {
		for i := 0; i < len(subscribersCallback); i++ {
			go subscribersCallback[i](message)
		}
	}
}

// Subscribe 订阅消息，通过 channel 异步返回。buffer 指定 Message 的数量
func (m *MemoryMessageQueue) Subscribe(topic string, buffer int) <-chan Message {
	ch := make(chan Message, buffer)
	m.Lock()
	m.topics[topic] = append(m.topics[topic], ch)
	m.Unlock()
	return ch
}

// SubscribeCallback 订阅函数消息，通过 callback function 异步执行。
func (m *MemoryMessageQueue) SubscribeCallback(topic string, callback CallbackFunc) {
	m.Lock()
	m.callbacks[topic] = append(m.callbacks[topic], callback)
	m.Unlock()
}

// Unsubscribe 取消订阅 channel 消息，sub 传入 Subscribe 返回的 channel，用于辨识取消哪个订阅者
func (m *MemoryMessageQueue) Unsubscribe(topic string, sub <-chan Message) {
	m.Lock()
	defer m.Unlock()

	subscribers, ok := m.topics[topic]
	if !ok {
		return
	}

	var newSubs []chan Message
	for _, subscriber := range subscribers {
		if subscriber == sub {
			continue
		}
		newSubs = append(newSubs, subscriber)
	}

	m.topics[topic] = newSubs
}
