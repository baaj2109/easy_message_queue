package main

import (
	"context"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	DelayTime int // desc 延遲時間
	Body      string
}

type MessageQueue struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	ExchangeName string
	RouteKey     string
	QueueName    string
}

type Consumer func(amqp.Delivery)

// new rabbitmq
func NewRabbitMQ(exchange, route, queue string) MessageQueue {
	messageQueue := MessageQueue{
		ExchangeName: exchange,
		RouteKey:     route,
		QueueName:    queue,
	}
	// connect to rabbitmq
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	// defer conn.Close()
	messageQueue.conn = conn
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	// defer ch.Close()
	messageQueue.ch = ch

	messageQueue.declareExchange(exchange, nil)
	return messageQueue
}

// send message to queue
func (mq *MessageQueue) Send(message Message) {
	err := mq.ch.PublishWithContext(
		context.Background(),
		mq.ExchangeName,
		mq.RouteKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message.Body),
		},
	)
	if err != nil {
		panic(err)
	}
}

// send delay message
func (mq *MessageQueue) SendDelay(message Message) {
	delayQueueName := mq.QueueName + "_delay_" + strconv.Itoa(message.DelayTime)
	delayRouteKey := mq.RouteKey + "_delay_" + strconv.Itoa(message.DelayTime)

	// declare delay queue
	dq := mq.declareQueue(
		delayQueueName,
		amqp.Table{
			"x-dead-letter-exchange":    mq.ExchangeName,
			"x-dead-letter-routing-key": mq.RouteKey,
		},
	)
	// bind queue to exchange
	mq.bindQueue(dq.Name, delayRouteKey, mq.ExchangeName)

	// send message to delay queue, after expired, send to original queue by auto route
	err := mq.ch.PublishWithContext(
		context.Background(),
		mq.ExchangeName,
		delayRouteKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message.Body),
			Expiration:  strconv.Itoa(message.DelayTime * 1000),
		},
	)
	if err != nil {
		panic(err)
	}
}

// consume message
func (mq *MessageQueue) Consume(fn Consumer) {
	// declare queue
	q := mq.declareQueue(mq.QueueName, nil)
	// bind queue to exchange
	mq.bindQueue(q.Name, mq.RouteKey, mq.ExchangeName)

	// set qos
	err := mq.ch.Qos(1, 0, false)
	if err != nil {
		panic(err)
	}
	msgs, err := mq.ch.Consume(
		q.Name, // queue name
		"",     // consumer
		false,  //auto-ack
		false,  //exclusive
		false,  // no-local
		false,  //no-wait
		nil,    //args
	)
	if err != nil {
		panic(err)
	}
	go func() {
		for d := range msgs {
			fn(d)
			d.Ack(false)
		}
	}()
}

func (mq *MessageQueue) Close() {
	mq.ch.Close()
	mq.conn.Close()
}

func (mq *MessageQueue) declareQueue(name string, args amqp.Table) amqp.Queue {
	q, err := mq.ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		panic("failed to declare a delay_queue" + err.Error())
	}
	return q
}

func (mq *MessageQueue) declareExchange(exchange string, args amqp.Table) {
	err := mq.ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		panic("failed to declare an exchange " + err.Error())
	}
}

func (mq *MessageQueue) bindQueue(queue, routekey, exchange string) {
	err := mq.ch.QueueBind(
		queue,
		routekey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		panic("failed to bind a queue" + err.Error())
	}
}

// / consume message
func RegisterRabbitMQConsumer() {
	// create new connect
	rabbit := NewRabbitMQ("myExchanger", "myRoute", "myQueue")
	// 一般來說消費者不會關閉，會常駐進程消費訊息
	// defer rabbit.Close()
	rabbit.Consume(func(d amqp.Delivery) {
		// logger.Info("rabbit",zap.String("rabbitmq", string(d.Body)))
	})

	//send message
	rabbit.Send(Message{Body: "normal message"})
	rabbit.SendDelay(Message{Body: "delay message"})

}
