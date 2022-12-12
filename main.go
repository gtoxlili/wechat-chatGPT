package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"wx-ChatGPT/chatGPT"
	"wx-ChatGPT/convert"
	"wx-ChatGPT/util"
)

const wxToken = "" // 这里填微信开发平台里设置的 Token

var reqGroup singleflight.Group

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(util.DefaultLogFormatter())
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger:  log.StandardLogger(),
			NoColor: runtime.GOOS == "windows",
		}))
	r.Use(middleware.Recoverer)

	// 微信接入校验
	r.Get("/weChatGPT", wechatCheck)
	// 微信消息处理
	r.Post("/weChatGPT", wechatMsgReceive)

	l, err := net.Listen("tcp", ":7458")
	if err != nil {
		log.Fatalln(err)
	}
	log.Infof("Server listening at %s", l.Addr())
	if err = http.Serve(l, r); err != nil {
		log.Fatalln(err)
	}
}

// 微信接入校验
func wechatCheck(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	signature := query.Get("signature")
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echostr := query.Get("echostr")

	// 校验
	if util.CheckSignature(signature, timestamp, nonce, wxToken) {
		render.PlainText(w, r, echostr)
		return
	}

	log.Errorln("微信接入校验失败")
}

// 微信消息处理
func wechatMsgReceive(w http.ResponseWriter, r *http.Request) {
	// 解析消息
	body, _ := io.ReadAll(r.Body)
	xmlMsg := convert.ToTextMsg(body)

	log.Infof("[消息接收] Type: %s, From: %s, MsgId: %d, Content: %s", xmlMsg.MsgType, xmlMsg.FromUserName, xmlMsg.MsgId, xmlMsg.Content)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// 回复消息
	replyMsg := ""

	// 关注公众号事件
	if xmlMsg.MsgType == "event" {
		if xmlMsg.Event == "unsubscribe" {
			chatGPT.DefaultGPT().DeleteUser(xmlMsg.FromUserName)
		}
		if xmlMsg.Event != "subscribe" {
			util.TodoEvent(w)
			return
		}
		replyMsg = ":) 感谢你发现了这里"
	} else if xmlMsg.MsgType == "text" {
		// 【收到不支持的消息类型，暂无法显示】
		if strings.Contains(xmlMsg.Content, "【收到不支持的消息类型，暂无法显示】") {
			util.TodoEvent(w)
			return
		}
		// 最多等待 15 s， 超时返回空值
		msg, err, _ := reqGroup.Do(strconv.FormatInt(xmlMsg.MsgId, 10), func() (interface{}, error) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			select {
			case msg := <-chatGPT.DefaultGPT().SendMsgChan(xmlMsg.Content, xmlMsg.FromUserName, ctx):
				return msg, nil
			case <-time.After(14*time.Second + 500*time.Millisecond):
				// 超时返回错误
				return "", fmt.Errorf("请求超时, MsgId: %d", xmlMsg.MsgId)
			}
		})
		if err != nil {
			log.Errorln(err)
			util.TodoEvent(w)
			return
		}
		replyMsg = msg.(string)
	} else {
		util.TodoEvent(w)
		return
	}

	textRes := &convert.TextRes{
		ToUserName:   xmlMsg.FromUserName,
		FromUserName: xmlMsg.ToUserName,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      replyMsg,
	}
	_, err := w.Write(textRes.ToXml())
	if err != nil {
		log.Errorln(err)
	}
}
