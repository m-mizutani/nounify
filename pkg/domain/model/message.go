package model

import "time"

type PubSubMessage struct {
	Data        []byte            `json:"data"`
	Attributes  map[string]string `json:"attributes"`
	MessageID   string            `json:"message_id"`
	PublishTime string            `json:"publish_time"`
}

type InputPubSub struct {
	Data        any               `json:"data"`
	Attributes  map[string]string `json:"attributes"`
	MessageID   string            `json:"message_id"`
	PublishTime time.Time         `json:"publish_time"`
}
