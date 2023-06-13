package broker

import "github.com/go-redis/redis"

type RedisClient struct {
	*redis.Client
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

func NewRedisClient(c RedisConfig) *RedisClient {
	return &RedisClient{redis.NewClient(&redis.Options{
		Addr:     c.Address,
		Password: c.Password,
		DB:       c.DB,
	})}
}

// / publish using redis client
func (r *RedisClient) Publish(topic string, message interface{}) error {
	return r.Client.Publish(topic, message).Err()
}

// / scribe using redis
func (r *RedisClient) Subscribe(topic string) *redis.PubSub {
	return r.Client.Subscribe(topic)
}

// rpush using redis
func (r *RedisClient) RPush(key string, value interface{}) error {
	return r.Client.RPush(key, value).Err()
}

// rpop using redis
func (r *RedisClient) RPop(key string) (string, error) {
	return r.Client.RPop(key).Result()
}

// lpush using redis
func (r *RedisClient) LPush(key string, value interface{}) error {
	return r.Client.LPush(key, value).Err()
}

// lpop using redis
func (r *RedisClient) LPop(key string) (string, error) {
	return r.Client.LPop(key).Result()
}
