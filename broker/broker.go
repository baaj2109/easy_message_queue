package broker

import "strings"

func NewBrokerClient(addr string, config interface{}) (Client interface{}) {
	if strings.Contains(addr, "redis") {
		Client = NewRedisClient(config.(RedisConfig))
		// } else if strings.Contains(addr, "amqp") {
		// 	Client = NewPikaRabbitMQClient(addr)
	} else {
		Client = nil
	}

	return

}
