package main

import "testing"

func TestWebSocketMessage_MarshalJSON(t *testing.T) {
	wsMesg := WebSocketMessage{Action: Action("action"), Message: "message"}
	json, err := wsMesg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	if string(json) != `["action","message"]` {
		panic(`json is not equal "["action","message"]"`)
	}
	println(string(json))
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
	println(wsMsg.Action)
	println(wsMsg.Message)
}
