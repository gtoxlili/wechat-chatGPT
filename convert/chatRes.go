package convert

import "encoding/json"

type ChatRes struct {
	Message        ChatResMessage `json:"message"`
	ConversationId string         `json:"conversation_id"`
}

type ChatResMessage struct {
	Id      string            `json:"id"`
	Content ChatResMsgContent `json:"content"`
}

type ChatResMsgContent struct {
	Parts []string `json:"parts"`
}

func ToChatRes(body []byte) *ChatRes {
	var msg ChatRes
	err := json.Unmarshal(body, &msg)
	if err != nil {
		panic(err)
	}
	return &msg
}

func (msg *ChatRes) ToJson() []byte {
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}
