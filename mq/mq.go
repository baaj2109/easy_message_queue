package mq

import "encoding/gob"

// Message 对应单一消息的结构体
type Message struct {
	Topic   string      // Topic 用于标识消息的对应的主题，消费者可以通过 Topic 过滤消息
	Context interface{} // Context 用于存放消息的内容
}

// CallbackFunc 用于定义消息处理函数
type CallbackFunc func(Message)

// Pool 用于定义消息队列的接口，所有的消息队列都需要实现这个接口
type Pool interface {
	Publish(topic string, message Message)                 // Publish 发布消息
	Subscribe(topic string, buffer int) <-chan Message     // Subscribe 订阅消息
	SubscribeCallback(topic string, callback CallbackFunc) // SubscribeCallback 订阅函数消息
	Unsubscribe(topic string, channel <-chan Message)      // Unsubscribe 取消订阅
}

/*
Pool 具有四个要求实现的方法：
Publish(topic string, message Message) 发布消息，传入 Topic 标识和消息内容，系统会将其传给所有订阅者。
Subscribe(topic string, buffer int) <-chan Message 订阅消息，通过传入所需消息的 Topic 和所需消息的多少，得到一个 Go Channel。后续可以通过 Channel 获取信息。
SubscribeCallback(topic string, callback CallbackFunc) 注册回调函数。通过传入消息的 Topic 为其注册回调函数，通常适用于较为单一的队列任务（例如 开卡、邮件发送等）。
Unsubscribe(topic string, channel <-chan Message) 通过传入 Topic 和之前通过 Subscribe 获取的 Channel 取消注册，避免不必要的消息推送。
*/

func init() {
	gob.Register(&Message{})
}
