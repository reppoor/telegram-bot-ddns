package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"telegrambot/config"
	"telegrambot/internal/bot/keyboard"
	"telegrambot/internal/db"
	"telegrambot/internal/db/repository"
	"telegrambot/internal/services"
	"telegrambot/internal/utils"
)

// HandleCommand handleCommand 用于处理不同的命令
func HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, Config *config.Config) {

	ID := update.Message.From.ID                     //消息发送者ID
	FirstName := update.Message.From.FirstName       //消息发送者名字
	LastName := update.Message.From.LastName         //消息发送者姓氏
	UserName := update.Message.From.UserName         //消息发送者用户名
	LanguageCode := update.Message.From.LanguageCode //消息发送者语言设置
	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			messageText := fmt.Sprintf("您好，很高兴为您服务") // 格式化消息内容，使用 Markdown 格式
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			_, _ = bot.Send(msg)
			return
		case "id":
			// 格式化消息内容，使用 Markdown 格式
			messageText := fmt.Sprintf("用户ID: `%d`\n名字: `%s`\n姓氏: `%s`\n用户名: [%s](https://t.me/%s)\n语言设置: `%s`", ID, FirstName, LastName, UserName, UserName, LanguageCode) // 格式化消息内容，使用 Markdown 格式
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			_, _ = bot.Send(msg)
			return
		case "init":
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用此命令`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}
			messageText := fmt.Sprintf("`机器人正常初始化数据库...`") // 格式化消息内容，使用 Markdown 格式
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			// 保存机器人发送的消息返回结果
			sentMsg, err := bot.Send(msg)
			if err != nil {
				fmt.Printf("发送初始化消息失败: %v\n", err)
				return
			}
			db.ATInitDB()
			db.CloseDB()
			// 编辑消息内容
			messageText = "`机器人数据库正常初始化完成`" // 格式化消息内容，使用 Markdown 格式
			editMsg := tgbotapi.NewEditMessageText(
				sentMsg.Chat.ID,   // 聊天 ID
				sentMsg.MessageID, // 需要编辑的消息 ID
				messageText,       // 新的消息内容
			)
			editMsg.ParseMode = "Markdown"
			// 编辑消息
			_, _ = bot.Send(editMsg)
			return
		case "info":
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用此命令`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}
			db.InitDB()
			DomainInfo, err := repository.GetDomainInfo()
			if err != nil {
				fmt.Println(err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"数据库未查询到任何域名记录❌️") // 要编辑的消息的 ID
				// 发送消息
				_, err = bot.Send(msg)
				return
			}
			keyBoard := keyboard.GenerateMainMenuKeyboard(DomainInfo) //生成内联键盘
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "查询转发信息")
			msg.ReplyMarkup = keyBoard
			// 发送消息
			_, err = bot.Send(msg)
			return
		case "check":
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用此命令`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}
			services.ALLCheckTCPConnectivity(bot, update, true)
			return
		case "insert":
			fmt.Printf("插入命令\n")
			// 获取命令部分（例如 /insert）
			command := update.Message.Command()
			// 提取命令后面的部分（参数）
			params := strings.TrimSpace(update.Message.Text[len(command)+1:]) // 去掉 "/insert " 部分
			_, err := utils.ValidateFormat(params)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf(params + "\n")
			DomainInfo := strings.Split(params, "#")
			port, err := strconv.Atoi(DomainInfo[2])
			db.InitDB() //连接数据库
			info, err := repository.InsertDomainInfo(DomainInfo[0], DomainInfo[1], port, DomainInfo[3])
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(info)
			return
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "抱歉，我不识别这个命令。")
			_, _ = bot.Send(msg)
			return
		}
	}
	if update.Message.Text != "" {
		fmt.Println("收到文本消息")
	}
}
