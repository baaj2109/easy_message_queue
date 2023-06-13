package mq

var Global Pool

func Init() {
	Global = NewMemoryMQ()
}
