package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"net/url"
	"telegrambot/config"
	"telegrambot/internal/bot/handlers"
)

func TelegramApp() {
	// 加载配置文件
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	var bot *tgbotapi.BotAPI
	if Config.Network.Proxy != "" {
		// 创建支持代理的 HTTP 客户端
		proxyURL, err := url.Parse(Config.Network.Proxy)
		if err != nil {
			log.Fatalf("解析代理地址失败: %v", err)
		}

		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
		bot, err = tgbotapi.NewBotAPIWithClient(Config.Telegram.Token, Config.Telegram.ApiEndpoint+"/bot%s/%s", httpClient)
		if err != nil {
			log.Panic(err)
		}
	} else {
		bot, err = tgbotapi.NewBotAPI(Config.Telegram.Token)
		if err != nil {
			log.Panic(err)
		}
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60                   // 设置超时时间
	updates := bot.GetUpdatesChan(u) // 获取更新通道
	// 在这里可以安全使用 bot 变量
	log.Printf("已授权账户: %s", bot.Self.UserName)
	//轮询消息
	for update := range updates {

		// 异步处理回调查询
		if update.CallbackQuery != nil {
			go handlers.CallbackQuery(bot, update, Config)
			continue
		}

		// 仅处理包含消息的更新
		if update.Message != nil {
			go handlers.HandleCommand(bot, update, Config)
			continue
		}
	}
}
