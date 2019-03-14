// Package lark is the missing Lark bot SDK for Go.
//
// Example Usage:
//
//  import (
//      "code.byted.org/zhangwanlong/go-lark"
//  )
//
//  const (
//      botToken = "b-d6da0626-9b57-46d7-a7d9-fd250ac4ed67"
//      botChannel = "6595007999788670976"
//  )
//
//  func main() {
//      bot := lark.NewChatBot(botToken)
//      bot.PostText("hello, world", botChannel)
//  }
//
package lark
