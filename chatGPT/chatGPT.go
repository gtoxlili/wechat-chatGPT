package chatGPT

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
	"unsafe"
	"wx-ChatGPT/convert"
	"wx-ChatGPT/util"
)

var (
	DefaultGPT = newChatGPT()
	// 对于每个 wxOpenID 都有独立的 parentID 和 conversationId
	// 但是对于同一个 wxOpenID，每次请求都会使用同一个 parentID 和 conversationId
	userInfoMap = util.NewSyncMap[string, *userInfo]()
)

type ChatGPT struct {
	authorization string
	sessionToken  string
}

type userInfo struct {
	parentID       string
	conversationId interface{}
	ttl            time.Time
}

func newChatGPT() *ChatGPT {
	sessionToken, err := os.ReadFile("sessionToken")
	if err != nil {
		log.Fatalln(err)
	}
	gpt := &ChatGPT{
		sessionToken: *(*string)(unsafe.Pointer(&sessionToken)),
	}
	// // 每 10 分钟更新一次 sessionToken
	go func() {
		gpt.updateSessionToken()
		for range time.Tick(10 * time.Minute) {
			gpt.updateSessionToken()
		}
	}()
	return gpt
}

func (c *ChatGPT) updateSessionToken() {
	session, err := http.NewRequest("GET", "https://chat.openai.com/api/auth/session", nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	session.AddCookie(&http.Cookie{
		Name:  "__Secure-next-auth.session-token",
		Value: c.sessionToken,
	})
	session.AddCookie(&http.Cookie{
		Name:  "__Secure-next-auth.callback-url",
		Value: "https://chat.openai.com/",
	})
	session.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	resp, err := http.DefaultClient.Do(session)
	if err != nil {
		log.Errorln(err)
		return
	}
	defer resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "__Secure-next-auth.session-token" {
			c.sessionToken = cookie.Value
			_ = os.WriteFile("sessionToken", []byte(cookie.Value), 0644)
			log.Infoln("sessionToken 更新成功 , sessionToken =", cookie.Value)
			break
		}
	}
	var accessToken map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&accessToken)
	if err != nil {
		log.Errorln(err)
		return
	}
	c.authorization = accessToken["accessToken"].(string)
}

func (c *ChatGPT) DeleteUser(OpenID string) {
	userInfoMap.Delete(OpenID)
}

func (c *ChatGPT) SendMsgChan(msg, OpenID string, ctx context.Context) <-chan string {
	ch := make(chan string, 1)
	go func() {
		ch <- c.SendMsg(msg, OpenID, ctx)
	}()
	return ch
}

func (c *ChatGPT) SendMsg(msg, OpenID string, ctx context.Context) string {
	// 获取用户信息
	info, ok := userInfoMap.Load(OpenID)
	if !ok || info.ttl.Before(time.Now()) {
		log.Infof("用户 %s 启动新的对话", OpenID)
		info = &userInfo{
			parentID:       uuid.New().String(),
			conversationId: nil,
		}
		userInfoMap.Store(OpenID, info)
	} else {
		log.Infof("用户 %s 继续对话", OpenID)
	}
	info.ttl = time.Now().Add(5 * time.Minute)
	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://chat.openai.com/backend-api/conversation", convert.CreateChatReqBody(msg, info.parentID, info.conversationId))
	if err != nil {
		log.Errorln(err)
		return "服务器异常, 请稍后再试"
	}
	req.Header.Set("Host", "chat.openai.com")
	req.Header.Set("Authorization", "Bearer "+c.authorization)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Openai-Assistant-App-Id", "")
	req.Header.Set("Connection", "close")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://chat.openai.com/chat")

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
	info.conversationId = res.ConversationId
	info.parentID = res.Message.Id
	if len(res.Message.Content.Parts) > 0 {
		return res.Message.Content.Parts[0]
	} else {
		return ""
	}
}
