package mq

import (
	"errors"
	"sync"
	"time"
)

type Broker interface {
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
broadCast：這個屬於内部方法，作用是進行廣播，對推送的消息進行广播，保証每一個訂閱者都可以收到
setConditions：這裡是用来設置條件，條件就是消息队列的容量，這樣我们就可以控制消息队列的大小了
*/

type BrokerImpl struct {
	exit     chan bool
	capacity int

	topics       map[string][]chan interface{} // key： topic  value ： queue
	sync.RWMutex                               // 同步锁
}

/*

exit：也是一个通道，这个用来做关闭消息队列用的
capacity：即用来设置消息队列的容量
topics：这里使用一个map结构，key即是topic，其值则是一个切片，chan类型，这里这么做的原因是我们一个topic可以有多个订阅者，所以一个订阅者对应着一个通道
sync.RWMutex：读写锁，这里是为了防止并发情况下，数据的推送出现错误，所以采用加锁的方式进行保证
*/

func NewBroker() *BrokerImpl {
	return &BrokerImpl{
		exit:   make(chan bool),
		topics: make(map[string][]chan interface{}),
	}
}

func (b *BrokerImpl) setConditions(capacity int) {
	b.capacity = capacity
}

func (b *BrokerImpl) close() {
	select {
	case <-b.exit:
		return
	default:
		close(b.exit)
		b.Lock()
		b.topics = make(map[string][]chan interface{})
		b.Unlock()
	}
	return
}

func (b *BrokerImpl) publish(topic string, pub interface{}) error {
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

func (b *BrokerImpl) broadcast(msg interface{}, subscribers []chan interface{}) {
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

func (b *BrokerImpl) subscribe(topic string) (<-chan interface{}, error) {
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

func (b *BrokerImpl) unsubscribe(topic string, sub <-chan interface{}) error {
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
