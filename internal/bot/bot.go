package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"telegrambot/config"
	"telegrambot/internal/bot/handlers"
	"telegrambot/internal/services"
	"time"
)

func TelegramApp() {
	// 加载配置文件
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	var bot *tgbotapi.BotAPI
	// 假设 Config.Network.Proxy 是代理地址，Config.Network.EnableProxy 是是否启用代理
	if Config.Network.EnableProxy {
		var httpClient *http.Client
		proxyURL, err := url.Parse(Config.Network.Proxy)
		if err != nil {
			log.Fatalf("解析代理地址失败: %v", err)
		}

		// 判断代理类型，HTTP 或 SOCKS5
		if strings.HasPrefix(proxyURL.Scheme, "http") {
			fmt.Println("使用http代理建立telegram连接")
			// 如果是 HTTP 代理
			httpClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
			}
		} else if strings.HasPrefix(proxyURL.Scheme, "socks5") {
			fmt.Println("使用socks5代理建立telegram连接")
			// 如果是 SOCKS5 代理
			dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
			if err != nil {
				log.Fatalf("创建 SOCKS5 代理失败: %v", err)
			}

			// 包装 dialer.Dial 成一个支持 DialContext 的方法
			httpClient = &http.Client{
				Transport: &http.Transport{
					DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
						// 使用代理的 Dial 方法
						return dialer.Dial(network, address)
					},
				},
			}
		} else {
			log.Fatalf("不支持的代理类型: %s", proxyURL.Scheme)
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
	// 创建一个单独的 Goroutine 用于定时任务
	go func() {
		ticker := time.NewTicker(Config.Check.CheckTime * time.Minute)
		defer ticker.Stop() // 确保程序退出时停止Ticker
		for {
			select {
			case <-ticker.C:
				// 创建一个模拟的 Update 对象
				up := tgbotapi.Update{
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{
							ID: Config.Telegram.Id, // 指定目标 Chat ID
						},
					},
				}
				fmt.Println("定时检测任务启动")
				services.ALLCheckTCPConnectivity(bot, up, false)
			}
		}
	}()

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
