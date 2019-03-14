# go-lark

The missing Lark bot SDK for Go.

* [Join go-lark User Channel](http://10.8.127.18:8000/)
* [Awesome Bots](/awesome-bots.md)

## Features

* Notification bot & chat bot supported
* Send messages (Group, Mention, Mention All, Private)
* Full supports of Lark Open API
* Build message with MessageBuilder (a.k.a. MsgBuf)
* Easy to create incoming message hook
* Encryption and Token Verification supported
* Documentation & tests

## Installation

```shell
go get -u code.byted.org/zhangwanlong/go-lark
```

## Supported API

<details>
    <summary>Show all...</summary>
    <br>
    <ul>Notification Bot
        <li>PostNotification</li>
    </ul>
    <ul>Chat Bot
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
        <li>Not implemented yet.</li>
    </ul>
</details>

## Send Messages

[Examples](#examples)

### Notification Bot

A Notification Bot is created from a Lark Group, where we will get a Webhook URL. We can use the Webhook to send message to this group.

```go
import (
    "code.byted.org/zhangwanlong/go-lark"
)

func main() {
    bot := lark.NewNotificationBot("WEB HOOK URL")
    bot.PostNotification("go-lark", "example")
}
```

### Chat Bot

A Chat Bot is created from [Lark Open Platform](https://lark-open.bytedance.net/list), which has much more features. We can build interactive chat bot with it.

```go
import (
    "code.byted.org/zhangwanlong/go-lark"
)

func main() {
    bot := lark.NewChatBot("BOT TOKEN")
    bot.PostText("hello, world", "YOUR CHANNEL")
}
```

Rich messages should be post with `PostMessage` directly. MessageBuilder can help construct a message body.

## MessageBuilder

MessageBuilder makes it easy to build a message body, which can be passed to `PostMessage`. It decouples the process of send message and build message body.

```go
mb := lark.NewMsgBuffer(lark.MsgText)
msg := mb.Text("hello, world").Mention("6454030812462448910", "Test").Build()

// hello, world<at user_id="6454030812462448910">@Test</at>
```

MessageBuilder support all kinds of message, including `MsgText`, `MsgPost`, `MsgImage`, and `MsgShareCard`. Specific functions go with certain message types. Otherwise, it will not take effect.

`Title` function is only available with `MsgPost`, and `Image` with `MsgImage`, `ShareChat` with `MsgShareCard`.

## Error Handling

A go-lark API returns system error and network error with `error`. API errors is embedded in specific API responses. Each ChatBot's API response contains `Code`, `Ok`, and `Error` to indicate error code, whether there is an error, and error message. We provide `IsEvent` and `GetError` to handle API errors.

```go
resp, err := bot.PostText("hello, world", "YOUR CHANNEL")
if err != nil {
    fmt.Errorf("Error is %v", err)
}
// IsError
if lark.IsError(resp) {
    fmt.Errorf("Error is %v", lark.GetError(resp))
}
// Use GetError directly
err = lark.GetError(resp)
if err != nil {
    fmt.Errorf("Error is %v", err)
}
```

## Event

(Incoming Message) Event is used to build interactive bot with Lark.

We may create event for a Chat Bot and then bind event hook to it. A hook URL will be bound with the bot/event, which will be the callback URL. We need to serve and respond to this URL.

For verification, Lark server will POST a challenge to the hook URL. We have to serve the URL and respond to the `challenge`.

Then, it will bind successfully. And incoming messages will be post to the URL.

In earlier version of Lark, we need `AddEvent` and `AddEventHook` to create a hook. And now, it is deprecated. Event can be created at [Lark Open Platform](https://lark-open.bytedance.net/page/app) with web interface.

We may use [Event Handler](#event-handler) to respond to the messages. And here is a [SHORTCUT](#shortcut).

### SHORTCUT

If we'd not like to DIY here, let us fast-forward to add an event hook with [add-lark-hook](/add-lark-hook/README_EN.md).

## Event Handler

Event Handler helps to deal with challenge and incoming messages.

### Challenge

See [examples/event-hook-challenge](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-hook-challenge).

```go
done := make(chan int)
ev := lark.NewEventHandler()
ev.ServeEventChallenge("/", ":9875", done)
```

### Incoming Message

See [examples/reply-event](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/reply-event).

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

We also provide a Gin middleware to handle incoming message.

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

### Encryption & Security

Lark Open Platform provides two methods for security, AES Encryption and Token Verification, which can be enabled in the meantime.

#### AES Encryption

AES Encryption should be enabled on server challenge, then all incoming messages will be encrypted.

You may call `EnableEncryption` to switch on and set `EncryptKey`.

```go
ev := lark.NewEventHandler()
ev.EnableEncryption("YOUR EncryptKey HERE")
// Then the following messages will be decrypted automatically
```

#### Token Verification

Token Verification is a way to ensure that the message is sent from Lark Open Platform.

```go
ev := lark.NewEventHandler()
ev.EnableTokenVerification("YOUR VERIFICATION TOKEN HERE")
// Then the following messages will be verified automatically
```

If you allow non-platform message sending (e.g. sent from `PostEvent`), don't enable it.

## Event for Testing

Lark Open Platform does not offer a client to test incoming message. We have to send messages from real Lark client to test `ServeEvent`.

The hook URL is also production URL, no debug URL. It's inconvenient and weird!

So we make `PostEvent` and use it to simulate incoming message. Thus, we can debug and test everywhere.

Even so, we have to build the incoming message manually.

See [examples/event-forward](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-forward).

```go
message := lark.EventMessage{
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
lark.PostEvent("http://localhost:9875/", message)
```

Meanwhile, we may also use `PostEvent` to forward a message, which acts as a reverse proxy.

## [Examples](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples)

* [Notification Message](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/notification-message)
* [Group Message](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/group-message)
* [Private Message](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/private-message)
* [Event Challenge](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-hook-challenge)
* [Event Handling](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/reply-event)
* [AES Encrypt](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/aes-encrypt)
* [Event Forwarding](https://code.byted.org/zhangwanlong/go-lark/tree/master/examples/event-forward)

## FAQ

* Send message error "request ip not allow": Disable "IP Filter" in Lark Open Platform.

## Contributing

* If you have a question about using URL Auto Redirector, start a discussion in [go-lark User Channel](http://10.8.127.18:8000/).
* If you think you've found a bug with go-lark, please [File an Issue](https://code.byted.org/zhangwanlong/go-lark/issues/new).
* Merge Request is welcomed.

---

Copyright (c) David Zhang, 2019.

Logo is designed by [OPEN LOGOS](https://github.com/arasatasaygin/openlogos).
