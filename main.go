package main

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// connect to rabbitmq
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// create channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// create queue
	queueName := "myQueue"
	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable 持久性
		false,     // delete when unused 自動刪除
		false,     // exclusive 排他性
		false,     // no-wait 不阻塞
		nil,       // arguments
	)

	if err != nil {
		log.Fatal(err)
	}
	// send message to queue
	message := "hello"
	err = ch.PublishWithContext(
		context.Background(),
		"",        // exchange
		queueName, // routing key
		false,     // mandatory 強制性
		false,     // immediate 立即
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent message : %s\n", message)

	// create a delayed queue
	exchangeName := "delayedExchange"
	err = ch.ExchangeDeclare(
		exchangeName,        // name
		"x-delayed-message", // type 交換機類型。 注意這是 rabbitmq 插件 "rabbitmq_delayed_message_exchange" 的類型
		true,                // durable
		false,               // delete when unused
		false,               // internal
		false,               // no-wait
		map[string]interface{}{
			"x-delayed-message": "direct", // exchange 交換機的類型，direct / topic / fanout
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	// declare queue
	quereName := "delayedQueue"
	_, err = ch.QueueDeclare(
		quereName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		map[string]interface{}{
			"x-message-ttl":             5000, // 消息的 TTL, 毫秒
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "myQueue", //死訊息發送的目標對列
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	message = "hello delayed message"
	err = ch.PublishWithContext(
		context.Background(),
		exchangeName, // exchange
		"",           // queue Name
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent delayed message : %s\n", message)

}
