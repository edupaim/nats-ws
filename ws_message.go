package main

import (
	"encoding/json"
	"errors"
)

type WebSocketMessage struct {
	Action  Action
	Message string
}

type Action string

const (
	ActionSubscribe   = "subscribe"
	ActionUnsubscribe = "unsubscribe"
)

func (message *WebSocketMessage) MarshalJSON() ([]byte, error) {
	a := []interface{}{
		message.Action,
		message.Message,
	}
	return json.Marshal(a)
}

func (message *WebSocketMessage) UnmarshalJSON(content []byte) error {
	var a []interface{}
	err := json.Unmarshal(content, &a)
	if err != nil {
		return err
	}
	if len(a) != 2 {
		return errors.New("wrong json format")
	}
	action, ok := a[0].(string)
	if !ok {
		return errors.New("wrong action type")
	}
	msg, ok := a[1].(string)
	if !ok {
		return errors.New("wrong action type")
	}
	message.Action = Action(action)
	message.Message = msg
	return nil
}
