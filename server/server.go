package main

import (
	"github.com/byronzhu-haha/chat/server/cmd"
)

func main() {
	server := cmd.NewChatServer()
	server.Run()
}
