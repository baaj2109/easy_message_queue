package mq

import (
	"errors"
	"sync"
	"time"
)

type IBroker interface {
	publish(topic string, msg interface{}) error
	subscribe(topic string) (<-chan interface{}, error)
	unsubscribe(topic string, sub <-chan interface{}) error
	close()
	broadcast(msg interface{}, subscribers []chan interface{})
	setConditions(capacity int)
}

/*
publish：進行消息的推送，有两個参數即topic、msg，分别是訂閱的主題、要傳遞的消息
subscribe：消息的訂閱，傳入訂閱的主題，即可完成訂閱，并返回對應的channel通道用来接收數據
unsubscribe：取消訂閱，傳入訂閱的主題和對應的通道
close：這個的作用就是很明顯了，就是用来關閉消息队列的
broadCast：這個屬於内部方法，作用是進行廣播，對推送的消息進行廣播，保証每一個訂閱者都可以收到
setConditions：這裡是用来設置條件，條件就是消息队列的容量，這樣我们就可以控制消息對列的大小了
*/

type Broker struct {
	exit     chan bool
	capacity int

	topics       map[string][]chan interface{} // key： topic  value ： queue
	sync.RWMutex                               // 同步锁
}

/*

exit：也是一个通道，這個用来做關閉消息隊列用的
capacity：即用来設置消息隊列的容量
topics：這裡使用一个map結構，key即是topic，其值是一个切片，chan類型，這麼做的原因是我们一个topic可以有多個訂閱者，所以一个訂閱者對應著一個通道
sync.RWMutex：讀寫鎖，這裡是為了防止併發情况下，數據的推送出現錯誤，所以採用加鎖的方式進行保證
*/

func NewBroker() *Broker {
	return &Broker{
		exit:   make(chan bool),
		topics: make(map[string][]chan interface{}),
	}
}

func (b *Broker) setConditions(capacity int) {
	b.capacity = capacity
}

func (b *Broker) close() {
	select {
	case <-b.exit:
		return
	default:
		/// 關閉 channel
		close(b.exit)
		b.Lock()
		b.topics = make(map[string][]chan interface{})
		b.Unlock()
	}
}

func (b *Broker) publish(topic string, pub interface{}) error {
	select {
	case <-b.exit:
		return errors.New("broker closed")
	default:
	}

	b.RLock()
	subscribers, ok := b.topics[topic]
	b.RUnlock()
	if !ok {
		return nil
	}

	b.broadcast(pub, subscribers)
	return nil
}

func (b *Broker) broadcast(msg interface{}, subscribers []chan interface{}) {
	count := len(subscribers)
	concurrency := 1

	switch {
	case count > 1000:
		concurrency = 3
	case count > 100:
		concurrency = 2
	default:
		concurrency = 1
	}

	pub := func(start int) {
		//采用Timer 而不是使用time.After 原因：time.After会产生内存泄漏 在计时器触发之前，垃圾回收器不会回收Timer
		idleDuration := 5 * time.Millisecond
		idleTimeout := time.NewTimer(idleDuration)
		defer idleTimeout.Stop()
		for j := start; j < count; j += concurrency {
			if !idleTimeout.Stop() {
				select {
				case <-idleTimeout.C:
				default:
				}
			}
			idleTimeout.Reset(idleDuration)
			select {
			case subscribers[j] <- msg:
			case <-idleTimeout.C:
			case <-b.exit:
				return
			}
		}
	}
	for i := 0; i < concurrency; i++ {
		go pub(i)
	}
}

func (b *Broker) subscribe(topic string) (<-chan interface{}, error) {
	select {
	case <-b.exit:
		return nil, errors.New("broker closed")
	default:
	}

	ch := make(chan interface{}, b.capacity)
	b.Lock()
	b.topics[topic] = append(b.topics[topic], ch)
	b.Unlock()
	return ch, nil
}

func (b *Broker) unsubscribe(topic string, sub <-chan interface{}) error {
	select {
	case <-b.exit:
		return errors.New("broker closed")
	default:
	}

	b.RLock()
	subscribers, ok := b.topics[topic]
	b.RUnlock()

	if !ok {
		return nil
	}
	// delete subscriber
	b.Lock()
	var newSubs []chan interface{}
	for _, subscriber := range subscribers {
		if subscriber == sub {
			continue
		}
		newSubs = append(newSubs, subscriber)
	}

	b.topics[topic] = newSubs
	b.Unlock()
	return nil
}
