package chatGPT

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
	"wx-ChatGPT/chatGPT/handler"
	"wx-ChatGPT/util"
)

var (
	once       sync.Once
	defaultGPT *ChatGPT
	// 对于每个 wxOpenID 都有独立的 parentID 和 conversationId
	// 但是对于同一个 wxOpenID，每次请求都会使用同一个 parentID 和 conversationId
	userInfoMap = util.NewSyncMap[string, *handler.UserInfo]()
)

func DefaultGPT() *ChatGPT {
	once.Do(func() {
		defaultGPT = newChatGPT()
	})
	return defaultGPT
}

type ChatGPT struct {
	authorization string
	config        *util.Config
}

func newChatGPT() *ChatGPT {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln("初始化失败: ", err)
		}
	}()
	config := util.ReadConfig()
	gpt := &ChatGPT{
		config: config,
	}
	// 每 10 分钟更新一次 config.json
	gpt.updateSessionToken()
	go func() {
		for range time.Tick(10 * time.Minute) {
			gpt.updateSessionToken()
		}
	}()
	return gpt
}

func (c *ChatGPT) updateSessionToken() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorln("更新 sessionToken 失败 :", err)
		}
	}()
	session, err := http.NewRequest("GET", "https://chat.openai.com/api/auth/session", nil)
	if err != nil {
		panic(err)
	}
	session.AddCookie(&http.Cookie{
		Name:  "__Secure-next-auth.session-token",
		Value: c.config.SessionToken,
	})
	session.AddCookie(&http.Cookie{
		Name:  "cf_clearance",
		Value: c.config.CfClearance,
	})
	session.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := http.DefaultClient.Do(session)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "__Secure-next-auth.session-token" {
			c.config.SessionToken = cookie.Value
			util.SaveConfig(c.config)
			log.Infoln("配置更新成功, sessionToken = ", cookie.Value)
			break
		}
	}
	var accessToken map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&accessToken)
	if err != nil {
		panic(err)
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
	return info.SendMsg(ctx, c.authorization, c.config, msg)
}
