package chatGPT

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io"
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
	// // 每 5 分钟更新一次 sessionToken
	go func() {
		gpt.updateSessionToken()
		for range time.Tick(5 * time.Minute) {
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

func (c *ChatGPT) SendMsg(msg, OpenID string) string {
	// 获取用户信息
	info, ok := userInfoMap.Load(OpenID)
	if !ok || info.ttl.Before(time.Now()) {
		log.Infof("用户 %s 启动新的对话", OpenID)
		info = &userInfo{
			parentID:       uuid.New().String(),
			conversationId: nil,
			ttl:            time.Now().Add(5 * time.Minute),
		}
		userInfoMap.Store(OpenID, info)
	} else {
		log.Infof("用户 %s 继续对话", OpenID)
	}
	info.ttl = time.Now().Add(5 * time.Minute)
	// 发送请求
	req, err := http.NewRequest("POST", "https://chat.openai.com/backend-api/conversation", convert.CreateChatReqBody(msg, info.parentID, info.conversationId))
	if err != nil {
		log.Errorln(err)
		return "服务器异常, 请稍后再试"
	}
	req.Header.Set("Authorization", "Bearer "+c.authorization)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorln(err)
		return "服务器异常, 请稍后再试"
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
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
