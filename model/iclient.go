package model

type IClient interface {
	Id() string
	Notify(message Message) error
}
