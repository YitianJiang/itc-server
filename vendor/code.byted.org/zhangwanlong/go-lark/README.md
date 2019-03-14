# go-lark

一个简单、开发者友好的 Lark 机器人封装。

* [加入 go-lark 开发者交流群](http://10.8.127.18:8000/)
* [Awesome Bots](/awesome-bots.md)
* [English README](/README_EN.md)

## 功能

* 支持通知机器人和聊天机器人
* 发送各类消息（群发、at、at全员、私聊）
* 其它各类开放平台已有接口
* 消息体构造 MessageBuilder
* 一站式解决服务器 Challenge 和聊天消息响应，快速构建交互机器人
* 支持加密和校验
* 文档、测试覆盖

## 安装

```shell
go get -u code.byted.org/zhangwanlong/go-lark
```

## 支持接口列表

<details>
    <summary>展开全部</summary>
    <br>
    <ul>通知机器人
        <li>PostNotification</li>
    </ul>
    <ul>聊天机器人
        <li>PostText</li>
        <li>PostTextMention</li>
        <li>PostTextMentionAll</li>
        <li>PostMessage</li>
        <li>PostPrivateMessage</li>
        <li>PostContent</li>
        <li>PostShareChatCard</li>
        <li>PostImage</li>
        <li>UploadImage</li>
        <li>OpenChat</li>
        <li>GetChannelInfo</li>
        <li>GetUserInfoByEmail</li>
        <li>GetUserInfo</li>
        <li>GetBotList</li>
        <li>GetBotInfo</li>
        <li>GetChannelList</li>
        <li>JoinChannel</li>
        <li>LeaveChannel</li>
        <li>CreateChannel</li>
        <li>AddChannelMember</li>
        <li>DeleteChannelMember</li>
        <li><s>AddEvent</s></li>
        <li><s>AddEventHook</s></li>
        <li>ServeEventChallenge</li>
        <li>ServeEvent</li>
        <li>PostEvent</li>
    </ul>
    <ul>H5App
        <li>暂未实现</li>
    </ul>
</details>

## 发送消息

[代码使用实例](#使用实例)

### 通知机器人

通过 Lark 群添加，会提供一个 Webhook。通过 Webhook 发消息，只能发送消息到这个群中。

```go
import (
    "code.byted.org/zhangwanlong/go-lark"
)

func main() {
    bot := lark.NewNotificationBot("WEB HOOK URL")
    bot.PostNotification("go-lark", "example")
}
```

### 聊天机器人

通过 Lark [开放平台](https://lark-open.bytedance.net/list)创建，功能比较多。可用于实现与机器人交互。

```go
import (
    "code.byted.org/zhangwanlong/go-lark"
)

func main() {
    bot := lark.NewChatBot("BOT TOKEN")
    bot.PostText("hello, world", "YOUR CHANNEL")
}
```

发送复杂的消息需要调用 `PostMessage`，可以通过 MessageBuilder 构造复杂的消息体。

## MessageBuilder

通过 MessageBuilder，可以构建更加复杂的消息体，把消息体构造和发消息之间解耦。

```go
mb := lark.NewMsgBuffer(lark.MsgText)
msg := mb.Text("hello, world").Mention("6454030812462448910", "Test").Build()

// hello, world<at user_id="6454030812462448910">@Test</at>
```

MessageBuilder 支持不同类型的消息，除了 `MsgText` 还有 `MsgPost`, `MsgImage` 和 `MsgShareCard`。部分函数跟消息类型是强关联的，类型错误不会生效。

`Title` 函数只能在 `MsgPost` 下使用，同时 `Image` 只能作用于 `MsgImage`，`ShareChat` 只能作用于 `MsgShareCard`。

## 错误处理

系统类、网络类错误通过 error 返回，服务返回错误在 response 中。ChatBot 的所有 API Response 都有 `Code`, `Ok`, 和 `Error`，分别是错误码、是否有错、错误信息。我们提供了 `IsError` 和 `GetError` 用于处理 API Response 中的错误。

```go
resp, err := bot.PostText("hello, world", "YOUR CHANNEL")
if err != nil {
    fmt.Errorf("Error is %v", err)
}
// IsError
if lark.IsError(resp) {
    fmt.Errorf("Error is %v", lark.GetError(resp))
}
// 直接使用 GetError
err = lark.GetError(resp)
if err != nil {
    fmt.Errorf("Error is %v", err)
}
```

## 事件 Event

事件是 Lark 机器人用于实现机器人交互的机制，创建机器人后，可以给机器人添加 Event，然后给 Event 绑定 Hook。Hook 的 URL 就是回调 URL，机器人开发者需要响应这个回调 URL 实现自己的功能。

早期版本中，Lark 需要使用 `AddEvent` 和 `AddEventHook` 添加绑定事件，目前已经废弃。现在需要在 [Lark 开放平台](https://lark-open.bytedance.net/page/app)-事件监听中创建事件。

完成创建后，Lark 服务器会发起 Challenge，向绑定的 URL Post 一个带有 `challenge` 的请求。需要返回 `challenge` 以通过验证。

通过验证后，所有发给这个机器人的消息都会被转发给绑定的 URL。

我们可以通过 [Event Handler](#event-handler) 响应这些请求。对此，我们提供了一个[快捷方法](#快捷方法)直接完成响应。

### 快捷方法

如果不想在此处 DIY，可以直接在响应事件的机器上使用 [add-lark-hook](/add-lark-hook)。

## Event Handler

通过 Event Handler，可以快速构建聊天机器人的事件响应部分（接收消息并且回复）。Event Handler 简单用 Go 封装了接口，只要实现一个函数就可以处理消息内容。

### 响应 Challenge

参考实例：[examples/event-hook-challenge](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-hook-challenge)

```go
done := make(chan int)
ev := lark.NewEventHandler()
ev.ServeEventChallenge("/", ":9875", done)
```

### 响应消息

参考实例：[examples/reply-event](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/reply-event)

```go
ev := lark.NewEventHandler()
ev.ServeEvent("/", ":9875", func(eventMsg lark.EventMessage) error {
    msg := eventMsg.Event
    userInfo, _ := bot.GetUserInfo(msg.UserID)
    _, err := bot.PostMessageMention(msg.UserID, userInfo.Name, "pong", msg.ChatID)
    return err
}
```

#### Gin Middleware

同时，响应消息也可以使用 Gin Middleware 方式。

```go
import "code.byted.org/zhangwanlong/go-lark/middleware"

r := gin.Default()
r.Use(middleware.GinLarkMessageMiddleware())
r.POST("/", func(c *gin.Context) {
    if message, ok := c.Get(DefaultLarkMessageKey); ok {
        m := message.(lark.EventMessage)
        text := m.Event.Text
        // your awesome logic
    }
})
r.Run(":9875")
```

### 加密安全

Lark 开放平台目前有两种加密安全策略（可以同时启用），分别是 AES 加密和 Token 校验。

#### AES 加密

AES 加密需要在验证 Challenge 时就开启，此后所有收到的消息都会走 AES 加密。

如果要开启 AES 加密，只需要使用 `EnableEncryption` 开启并设置 `EncryptKey`：

```go
ev := lark.NewEventHandler()
ev.EnableEncryption("YOUR EncryptKey HERE")
// 开启后所有消息会被自动解密
```

#### Token 校验

Token 校验是为了验证消息来自 Lark 开放平台。它的用法比较简单：

```go
ev := lark.NewEventHandler()
ev.EnableTokenVerification("YOUR VERIFICATION TOKEN HERE")
// 开启后所有消息会被自动校验
```

如果你要允许非平台发送的消息（比如：使用 `PostEvent` 发送），那请不要开启。

## 测试消息事件

Lark 官方没有提供发消息工具，如果测试 `ServeEvent` 的话不得不在 Lark 上发消息，直接在“线上” URL 调试，很不方便。

我们加入了线下模拟消息事件的 `PostEvent`，通过它可以在任何地方进行调试。当然，模拟消息的包体需要自己构造。

参考实例：[examples/event-forward](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-forward)

```go
ev := lark.NewEventHandler()
message := ev.EventMessage{
    Timestamp: "",
    Token:     "",
    EventType: "event_callback",
    Event: lark.EventText{
        Type:          "message",
        ChatType:      "private",
        MsgType:       "text",
        UserID:        "6454030812462448910",
        ChatID:        "6579052343910727944",
        Text:          "tlb",
        Title:         "",
        OpenMessageID: "",
        ImageKey:      "",
        ImageURL:      "",
    },
}
ev.PostEvent("http://localhost:9875/", message)
```

同时，我们也可以使用 `PostEvent` 对原消息进行转发，对消息进行反向代理。

## [使用实例](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples)

* [通知发送](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/notification-message)
* [群消息发送](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/group-message)
* [私聊消息发送](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/private-message)
* [事件 Challenge](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-hook-challenge)
* [事件响应](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/reply-event)
* [AES 加密](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/aes-encrypt)
* [事件转发](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-forward)

## 常见问题

* 调用接口发消息报错 "request ip not allow"：在开发者后台取消“IP白名单”

## 贡献

* 如果在仔细阅读文档后有关于 go-lark 的问题，可以[加入 go-lark 开发者交流群](http://10.8.127.18:8000/)发起讨论。
* 如果在使用 go-lark 时遇到 Bug，请[提交 Issue](https://code.byted.org/zhangwanlong/go-lark/issues/new)。
* 欢迎通过 Merge Request 提交功能或 Bug 修复。

---

Copyright (c) David Zhang, 2019.

Logo is designed by [OPEN LOGOS](https://github.com/arasatasaygin/openlogos).
