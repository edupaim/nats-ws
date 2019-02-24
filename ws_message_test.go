package main

import (
	"testing"
)

func TestWebSocketMessage_MarshalJSON(t *testing.T) {
	wsMesg := WebSocketMessage{Action: Action("action"), Message: "message"}
	json, err := wsMesg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	if string(json) != `["action","message"]` {
		panic(`json is not equal "["action","message"]"`)
	}
}

func TestWebSocketMessage_UnmarshalJSON(t *testing.T) {
	var wsMsg WebSocketMessage
	err := wsMsg.UnmarshalJSON([]byte(`["action","message"]`))
	if err != nil {
		panic(err)
	}
	if wsMsg.Action != Action("action") {
		panic(`action is not equal "action"`)
	}
	if wsMsg.Message != "message" {
		panic(`message is not equal "message"`)
	}
}

func TestWebSocketMessage_UnmarshalJSON_WrongJsonFormat(t *testing.T) {
	var wsMsg WebSocketMessage
	err := wsMsg.UnmarshalJSON([]byte(`{"action":"message"}`))
	if err == nil {
		panic("want error")
	}

}

func TestWebSocketMessage_UnmarshalJSON_WrongJsonFormat_2(t *testing.T) {
	var wsMsg WebSocketMessage
	err := wsMsg.UnmarshalJSON([]byte(`["action","message","other"]`))
	if err.Error() != "wrong json format" {
		panic("want wrong json format error")
	}
}
