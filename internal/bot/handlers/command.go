package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
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
			fmt.Printf("start命令\n")
			messageText := fmt.Sprintf("您好，很高兴为您服务") // 格式化消息内容，使用 Markdown 格式
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			_, _ = bot.Send(msg)
			return
		case "id":
			fmt.Printf("id命令\n")
			// 格式化消息内容，使用 Markdown 格式
			messageText := fmt.Sprintf(
				"*👤 用户信息:*\n\n"+
					"*🆔 用户ID:* `%d`\n"+
					"*🧑 名字:* `%s`\n"+
					"*👨‍🦱 姓氏:* `%s`\n"+
					"*🔗 用户名:* [%s](https://t.me/%s)\n"+
					"*🌐 语言设置:* `%s`",
				ID, FirstName, LastName, UserName, UserName, LanguageCode)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			_, _ = bot.Send(msg)
			return
		case "init":
			fmt.Printf("init命令\n")
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用init命令`") // 格式化消息内容，使用 Markdown 格式
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
			fmt.Printf("info命令\n")
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用info命令`") // 格式化消息内容，使用 Markdown 格式
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
			fmt.Printf("check命令\n")
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用check命令`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}
			// 获取所有域名信息，假设按ID排序
			db.InitDB()
			_, err := repository.GetDomainInfo()
			if err != nil {
				fmt.Println("获取域名信息失败:", err)
				messageText := fmt.Sprintf("数据库未查询到任何域名记录❌️") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}
			services.ALLCheckTCPConnectivity(bot, update, true)
			return
		case "insert":
			fmt.Printf("insert命令\n")
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用insert命令`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}
			// 获取命令部分（例如 /insert）
			command := update.Message.Command()
			// 提取命令后面的部分（参数）
			params := strings.TrimSpace(update.Message.Text[len(command)+1:]) // 去掉 "/insert " 部分
			// 参数格式验证
			_, err := utils.ValidateFormat(params)
			if err != nil {
				messageText := fmt.Sprintf(
					"*📌 请参考以下格式:*\n\n"+
						"*📝 格式说明:*\n"+
						"`主域名#转发域名#转发端口#运营商`\n\n"+
						"*📍 单条记录示例:*\n"+
						"`www.baidu.com#www.hao123.com#7890#运营商`\n\n"+
						"*📦 批量记录示例（转发域名用 `|` 分隔）:*\n"+
						"`www.baidu.com#www.hao123.com|www.4399.com#7890#运营商A|运营商B`\n\n"+
						"*❗️检测到的非法格式:*\n"+
						"`%s`",
					err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				fmt.Println(err)
				return
			}
			// 解析参数
			fmt.Printf(params + "\n")
			parts := strings.Split(params, "#")
			// 获取主要域名和需要遍历的域名列表
			primaryDomain := strings.TrimSpace(parts[0]) // 主要域名
			domainList := strings.Split(parts[1], "|")   // 遍历的域名
			port, err := strconv.Atoi(parts[2])          // 端口号
			if err != nil {
				messageText := "*端口号格式错误，请输入数字*"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}

			// 处理运营商字段
			operatorList := strings.Split(parts[3], "|")

			// 检查域名和运营商是否一一对应
			if len(domainList) != len(operatorList) {
				messageText := "*格式错误:* `域名列表和运营商列表数量不匹配，请检查`\n例如: \n`www.baidu.com#www.hao123.com|www.4399.com#7890#运营商A|运营商B`"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}

			// 初始化数据库连接
			db.InitDB()

			// 插入域名和对应的运营商
			var successCount, failCount int
			for i, domain := range domainList {
				domain = strings.TrimSpace(domain)
				operator := strings.TrimSpace(operatorList[i])
				if domain == "" {
					continue
				}
				if operator == "" {
					operator = "未备注" // 默认值
				}

				info, err := repository.InsertDomainInfo(primaryDomain, domain, port, operator)
				if err != nil {
					fmt.Printf("插入域名 %s 失败: %v\n", domain, err)
					failCount++
				} else {
					fmt.Printf("插入域名 %s 成功: %v\n", domain, info)
					successCount++
				}
			}

			// 返回操作结果
			messageText := fmt.Sprintf("插入完成✅️\n成功: %d 条\n失败: %d 条", successCount, failCount)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			_, _ = bot.Send(msg)
			return
		case "version":
			fmt.Printf("version命令\n")
			v := services.Version()
			messageText := fmt.Sprintf(v) // 格式化消息内容，使用 Markdown 格式
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			_, _ = bot.Send(msg)
		case "parse":
			fmt.Println("parse命令")
			if ID != Config.Telegram.Id {
				messageText := "*🚫 您无权限使用该命令*"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}

			db.InitDB()

			allDomains, err := repository.GetDomainInfo()
			if err != nil {
				fmt.Println("获取域名信息失败:", err)
				messageText := "*❌ 数据库未查询到任何域名记录*"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}

			var orderedDomains []string
			domainSet := make(map[string]struct{})
			for _, d := range allDomains {
				if _, exists := domainSet[d.Domain]; !exists {
					domainSet[d.Domain] = struct{}{}
					orderedDomains = append(orderedDomains, d.Domain)
				}
			}

			var domainInfoList []string
			for _, domainName := range orderedDomains {
				info, err := services.GetCloudflareDomainInfo(domainName)
				if err != nil {
					log.Printf("获取域名 %s 信息失败: %v\n", domainName, err)
					continue
				}

				recordTypeText := "CNAME记录"
				if info.RecordType {
					recordTypeText = "A记录"
				}

				infoString := fmt.Sprintf(
					"🌐 *域名:* `%s`\n🔀 *转发域:* `%s`\n✏️ *记录类型:* `%s`\n📥 *IP:* `%s`\n🏢 *运营商:* `%s`",
					info.Domain, info.ForwardingDomain, recordTypeText, info.IP, info.ISP,
				)
				domainInfoList = append(domainInfoList, infoString)
			}

			finalSentence := strings.Join(domainInfoList, "\n\n──────────────\n\n")
			if finalSentence == "" {
				finalSentence = "_⚠️ 没有可用的域名解析记录_"
			}

			messageText := "*📦 当前 Cloudflare 解析情况:*\n\n" + finalSentence
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			_, _ = bot.Send(msg)
		case "getip":
			fmt.Println("getip命令")
			if ID != Config.Telegram.Id {
				messageText := "*🚫 无权限使用 getIp 命令*"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}

			messageText := "*📡 开始处理域名解析*\n\n" +
				"处理进度: `0%%`\n" +
				"_正在写入转发 IP，请稍候..._"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			msg.ParseMode = "Markdown"
			sentMessage, _ := bot.Send(msg)

			// 初始化数据库
			db.InitDB()

			// 获取域名数据
			Domains, err := repository.GetALLDomain()
			if err != nil {
				fmt.Println("获取域名信息失败:", err)
				messageText = "*❗️ 获取域名信息失败!*"
				msg := tgbotapi.NewEditMessageText(update.Message.Chat.ID, sentMessage.MessageID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}

			if Domains == nil {
				log.Println("没有任何域名数据")
				messageText = "*⚠️ 没有任何域名数据可处理*"
				msg := tgbotapi.NewEditMessageText(update.Message.Chat.ID, sentMessage.MessageID, messageText)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				return
			}

			totalDomains := len(Domains)

			// 遍历并处理域名
			for i, domain := range Domains {
				newIP, err := services.ResolveDomainToIP(domain.ForwardingDomain)
				if err != nil {
					messageText := fmt.Sprintf("*❌ 域名解析失败*\n`%s` 无法解析 IP", domain.ForwardingDomain)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					continue
				}

				idStr := fmt.Sprintf("%d", domain.ID)
				_, err = repository.UpdateDomainIp(idStr, newIP)
				if err != nil {
					messageText := fmt.Sprintf("*⚠️ 数据库更新失败*\n域名: `%s`\n目标 IP: `%s`", domain.ForwardingDomain, newIP)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					continue
				}

				progress := int(float64(i+1) / float64(totalDomains) * 100)

				if progress == 100 {
					messageText = fmt.Sprintf(
						"*✅ 所有域名处理完成*\n\n"+
							"共处理域名: *%d*\n"+
							"最后一项:\n"+
							"🌐 `%s`\n"+
							"📥 IP: `%s`",
						totalDomains, domain.ForwardingDomain, newIP)
				} else {
					messageText = fmt.Sprintf(
						"*🔁 处理进度:* `%d%%`\n"+
							"*🌐 域名:* `%s`\n"+
							"*📥 新转发 IP:* `%s`\n"+
							"✅ 更新成功",
						progress, domain.ForwardingDomain, newIP)
				}

				editProgressMsg := tgbotapi.NewEditMessageText(update.Message.Chat.ID, sentMessage.MessageID, messageText)
				editProgressMsg.ParseMode = "Markdown"
				_, _ = bot.Send(editProgressMsg)
			}
		case "delete":
			fmt.Printf("delete命令\n")
			if ID != Config.Telegram.Id {
				messageText := fmt.Sprintf("`您无法使用delete命令`") // 格式化消息内容，使用 Markdown 格式
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
			keyBoard := keyboard.GenerateMainMenuDeleteKeyboard(DomainInfo) //生成内联键盘
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "请选择删除的转发记录\n"+
				"✅️=删除\n"+
				"🚫=不删")
			msg.ReplyMarkup = keyBoard
			// 发送消息
			_, err = bot.Send(msg)
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
