package handler

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"unsafe"
	"wx-ChatGPT/convert"
	"wx-ChatGPT/util"
)

var baseHeader = map[string]string{
	"User-Agent":   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	"Accept":       "text/event-stream",
	"Content-Type": "application/json",
	"Connection":   "close",
}

type UserInfo struct {
	parentID       string
	conversationId interface{}
	TTL            time.Time
}

func NewUserInfo() *UserInfo {
	return &UserInfo{
		parentID:       uuid.New().String(),
		conversationId: nil,
	}
}

func (user *UserInfo) SendMsg(ctx context.Context, authorization, msg, cfClearance string) string {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://chat.openai.com/backend-api/conversation", convert.CreateChatReqBody(msg, user.parentID, user.conversationId))
	if err != nil {
		log.Errorln(err)
		return "服务器异常, 请稍后再试"
	}
	for k, v := range baseHeader {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authorization)
	req.AddCookie(&http.Cookie{
		Name:  "cf_clearance",
		Value: cfClearance,
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorln(err)
		return "服务器异常, 请稍后再试"
	}
	defer resp.Body.Close()
	bodyBytes, err := util.ReadWithCtx(ctx, resp.Body)
	defer util.PutBytes(bodyBytes)
	if err != nil {
		log.Errorln(err)
		return "服务器异常, 请稍后再试"
	}
	line := bytes.Split(bodyBytes, []byte("\n\n"))
	if len(line) < 2 {
		log.Errorln(*(*string)(unsafe.Pointer(&bodyBytes)))
		return "服务器异常, 请稍后再试"
	}
	endBlock := line[len(line)-3][6:]
	res := convert.ToChatRes(endBlock)
	user.conversationId = res.ConversationId
	user.parentID = res.Message.Id
	if len(res.Message.Content.Parts) > 0 {
		return res.Message.Content.Parts[0]
	} else {
		return ""
	}
}
