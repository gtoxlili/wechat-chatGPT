## wechat-chatGPT

具有微信公众号被动回复用户消息功能的 ChatGPTBot 实现

### 食用指南
1. 在 main.go 文件中填入，微信公众平台中设置的 Token
2. 在 sessionToken 文件中填入`chat.openai.com` 里 Cookie 中的 __Secure-next-auth.session-token
3. 编译项目，并部署到服务器中 默认监听本机 127.0.0.1:7458, 请自行通过 Nginx 或 Caddy 等反向代理工具进行转发
4. 在微信公众平台中设置服务器地址为你的服务器地址

### 编译命令
```shell
GOOS=linux GOARCH=amd64 GOARM= GOMIPS= \
CGO_ENABLED=0 \                                                   
go build -trimpath -o ./dist/weChatGPT \                          
-ldflags "-w -s -buildid="
```
### 注意事项
sessionToken 文件请放置与可执行文件同一目录下，并且只保留__Secure-next-auth.session-token的内容
### 效果图
![效果图](\img\photo.jpg)

### 其他
> ~~这其实是一篇没什么用的README~~  
> 
> 由于微信公众号的 [5s限制](https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Passive_user_reply_message.html)，虽然本项目已经通过技术将这个限制提升至了 15s,
> 但绝大多数情况下通过逆向得到的ChatGPT接口的相应速率都超过了这个时间限制。  
> 
> 故本 Bot 几乎无法正常工作，可能以后等 ChatGPT 的正式接口出来，会重构本项目的代码。