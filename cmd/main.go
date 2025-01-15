package main

import (
	"fmt"
	"telegrambot/internal/bot"
)

func main() {
	fmt.Print("当前BOT运行版本:1.0.2")
	bot.TelegramApp() // APP入口
}
