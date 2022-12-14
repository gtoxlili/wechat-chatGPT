## WeChat-chatGPT

具有微信公众号被动回复用户消息功能的 ChatGPTBot 实现

### 食用指南

1. 在 config.json 文件中填入`chat.openai.com` 里 Cookie 中的 __Secure-next-auth.session-token 与 cf_clearance
2. 编译项目，注意在编译时将 `$(Token)` 替换为你的微信公众号 Token
3. 部署到服务器中 默认监听本机 127.0.0.1:7458, 请自行通过 Nginx 或 Caddy 等反向代理工具进行转发
3. 在微信公众平台中设置服务器地址为你的反向代理地址或域名地址，与微信公众号绑定的路由为 `/weChatGPT`

### 编译命令

```shell
GOOS=linux GOARCH=amd64 GOARM= GOMIPS= \
CGO_ENABLED=0 \                                                   
go build -trimpath -o ./dist/weChatGPT \                          
-ldflags "-X 'main.wxToken=$(Token)' -w -s -buildid="
```

### 注意事项

1. `config.json` 文件请放置与可执行文件同一目录下
2. `cf_clearance` 可用于绕过 `Cloudflare` 的防火墙，但请保证获取 `cf_clearance` 时的 UA 与 IP 与项目实际运行时一致 (本项目默认
   UA 为 `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36`)

### 效果图

![](https://github.com/gtoxlili/wechat-chatGPT/blob/master/img/screenshot.jpg?raw=true)

### 关于如何在服务器上获取 `cf_clearance`

以下流程以Linode的Ubuntu镜像为例：

1、在服务器上安装[vvanglro/cf-clearance](https://github.com/vvanglro/cf-clearance)：
```shell
pip install git+https://github.com/vvanglro/cf-clearance.git@main
```
该项目需要`playwright`的支持，最新版Ubuntu镜像貌似是自带的（较老的Ubuntu版本需要自行安装），但还需要装一些依赖内容：
```shell
playwright install
apt-get install libatk1.0-0 libatk-bridge2.0-0 libcups2 libatspi2.0-0 libxcomposite1 libxdamage1 libxfixes3 libxrandr2 libgbm1 libxkbcommon0 libpango-1.0-0 libcairo2 libasound2
```

3、安装xvfb，用来启动虚拟GUI
```shell
sudo apt-get install xvfb
```

4、然后在服务器上创建一个`get_cf.py`脚本，内容如下：

```python
from playwright.sync_api import sync_playwright
from cf_clearance import sync_cf_retry, sync_stealth

with sync_playwright() as p:
    browser = p.chromium.launch(headless=False)
    page = browser.new_page()
    sync_stealth(page)
    page.goto('https://chat.openai.com/chat')
    res = sync_cf_retry(page)
    if res:
        cppkies = page.context.cookies()
        for cookie in cppkies:
            if cookie.get('name') == 'cf_clearance':
                print(cookie.get('value'))
        ua = page.evaluate('() => {return navigator.userAgent}')
        print(ua)
    else:
        print("fail")
    browser.close()
```

5、需要进入服务器的交互模式，从Linode控制台以Glish方式登录。进入到存放`get_cf.py`脚本的位置，运行以下命令：

```shell
xvfb-run python3 get_cf.py > cf.txt
```

成功后，退出Glish终端（因为无法在里面复制内容）

6、打开生成的cf.txt，里面保存的就是`cf_clearance`和`user-agent` ,复制到`config.json`中即可。

### 其他

> ~~这其实是一篇没什么用的README~~
>
>
> 由于微信公众号的 [5s限制](https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Passive_user_reply_message.html)
，虽然本项目已经通过技术将这个限制提升至了 15s,
> 但绝大多数情况下通过逆向得到的ChatGPT接口的相应速率都超过了这个时间限制。
>
> 故本 Bot 几乎无法正常工作，可能以后等 ChatGPT 的正式接口出来，会重构本项目的代码。
