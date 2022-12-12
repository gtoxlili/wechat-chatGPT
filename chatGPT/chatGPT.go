package chatGPT

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
	"unsafe"
	"wx-ChatGPT/chatGPT/handler"
	"wx-ChatGPT/util"
)

var (
	DefaultGPT = newChatGPT()
	// 对于每个 wxOpenID 都有独立的 parentID 和 conversationId
	// 但是对于同一个 wxOpenID，每次请求都会使用同一个 parentID 和 conversationId
	userInfoMap = util.NewSyncMap[string, *handler.UserInfo]()
)

type ChatGPT struct {
	authorization string
	sessionToken  string
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
	if !ok || info.TTL.Before(time.Now()) {
		log.Infof("用户 %s 启动新的对话", OpenID)
		info = handler.NewUserInfo()
		userInfoMap.Store(OpenID, info)
	} else {
		log.Infof("用户 %s 继续对话", OpenID)
	}
	info.TTL = time.Now().Add(5 * time.Minute)
	// 发送请求
	return info.SendMsg(ctx, c.authorization, msg)
}
