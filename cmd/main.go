package main

import (
	"fmt"
	"telegrambot/internal/bot"
	"telegrambot/internal/services"
)

func main() {
	v := services.Version()
	fmt.Print(v + "\n")
	bot.TelegramApp() // APP入口
}
