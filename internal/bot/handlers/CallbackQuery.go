package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
	"telegrambot/config"
	"telegrambot/internal/bot/Keyboard"
	"telegrambot/internal/db"
	"telegrambot/internal/db/repository"
	"telegrambot/internal/services"
)

func CallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, Config *config.Config) {

	data := update.CallbackQuery.Data
	// 将回调数据按 '-' 分隔，判断菜单层级
	levels := strings.Split(data, "-")

	switch len(levels) {
	case 1:
		db.InitDB() //连接数据库
		DomainInfo, err := repository.GetDomainIDInfo(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		ID := DomainInfo.ID
		Domain := DomainInfo.Domain
		ForwardingDomain := DomainInfo.ForwardingDomain
		IP := DomainInfo.IP
		Port := DomainInfo.Port
		ISP := DomainInfo.ISP
		Ban := DomainInfo.Ban
		// 格式化消息内容，使用 Markdown 格式
		messageText := fmt.Sprintf(
			"ID: `%d`\n域名: `%s`\n转发域名: `%s`\nIP: `%s`\n端口: `%d`\n运营商: `%s`\nIsBan: `%t`",
			ID, Domain, ForwardingDomain, IP, Port, ISP, Ban,
		) // 格式化消息内容，使用 Markdown 格式
		fmt.Println(messageText)
		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
			update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
			messageText,                            // 新的消息文本
		)
		msg.ParseMode = "Markdown"
		// 创建按钮
		msg.ReplyMarkup = Keyboard.GenerateSubMenuKeyboard(ID, Ban)
		_, err = bot.Send(msg)
		fmt.Println("当前是1级菜单")
	case 2:
		if len(levels) > 1 {
			ID := levels[0]
			action := levels[1]

			switch action {
			case "del":
				// 处理删除操作
				fmt.Println("执行删除操作, ID:", ID)
				messageText := fmt.Sprintf("`正在删除该条记录...`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				_, err := repository.DeleteDomainByID(data)
				if err != nil {
					messageText = fmt.Sprintf("`删除失败❌️`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				db.InitDB()
				DomainInfo, err := repository.GetDomainInfo()
				if err != nil {
					fmt.Println(err)
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID, // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID,
						"数据库未查询到任何域名记录❌️") // 要编辑的消息的 ID
					// 发送消息
					_, err = bot.Send(msg)
					return
				}
				keyboard := Keyboard.GenerateMainMenuKeyboard(DomainInfo) //生成内联键盘
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID, // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID,
					"记录删除成功✅️") // 要编辑的消息的 ID
				msg.ReplyMarkup = &keyboard
				// 发送消息
				_, err = bot.Send(msg)
			case "parse":
				// 格式化消息内容，使用 Markdown 格式
				messageText := fmt.Sprintf("`正在解析DNS记录...`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				// 处理解析操作
				fmt.Println("执行解析操作, ID:", ID)
				db.InitDB() //连接数据库
				DomainInfo, err := repository.GetDomainIDInfo(data)
				if err != nil {
					fmt.Println("查询数据库失败", err)
					// 格式化消息内容，使用 Markdown 格式
					messageText = fmt.Sprintf("`查询数据库失败`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				newIP, err := services.ResolveDomainToIP(DomainInfo.ForwardingDomain) //获取转发IP
				if err != nil {
					fmt.Println("获取IP失败", err)
					// 格式化消息内容，使用 Markdown 格式
					messageText = fmt.Sprintf("`获取IP失败`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				_, err = services.UpdateARecord(DomainInfo.Domain, newIP)
				if err != nil {
					fmt.Println("更新域名A记录失败", err)
					// 格式化消息内容，使用 Markdown 格式
					messageText = fmt.Sprintf("`更新域名A记录失败`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				newDomainIp, err := repository.UpdateDomainIp(data, newIP)
				if err != nil {
					fmt.Println("更新数据库IP失败", err)
					return
				}
				ID := newDomainIp.ID
				Domain := newDomainIp.Domain
				ForwardingDomain := newDomainIp.ForwardingDomain
				IP := newDomainIp.IP
				Port := newDomainIp.Port
				ISP := newDomainIp.ISP
				Ban := newDomainIp.Ban
				// 格式化消息内容，使用 Markdown 格式
				messageText = fmt.Sprintf(
					"*解析成功*✅\nID: `%d`\n域名: `%s`\n转发域名: `%s`\nIP: `%s`\n端口: `%d`\n运营商: `%s`\nIsBan: `%t`",
					ID, Domain, ForwardingDomain, IP, Port, ISP, Ban,
				) // 格式化消息内容，使用 Markdown 格式
				fmt.Println(messageText)
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				// 创建按钮
				msg.ReplyMarkup = Keyboard.GenerateSubMenuKeyboard(ID, Ban)
				_, err = bot.Send(msg)
			case "checkAndParse":
				// 检测连通性并解析记录
				fmt.Println("执行检测连通性并解析记录, ID:", ID)
				// 格式化消息内容，使用 Markdown 格式
				messageText := fmt.Sprintf("`正在检测连通性...`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				db.InitDB() //连接数据库
				DomainInfo, err := repository.GetDomainIDInfo(data)
				if err != nil {
					fmt.Println("查询数据库失败", err)
					// 格式化消息内容，使用 Markdown 格式
					messageText = fmt.Sprintf("`查询数据库失败`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				newIP, err := services.ResolveDomainToIP(DomainInfo.ForwardingDomain) //获取IP
				if err != nil {
					fmt.Println("获取IP失败", err)
					// 格式化消息内容，使用 Markdown 格式
					messageText = fmt.Sprintf("`获取IP失败`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				if !services.CheckTCPConnectivity(newIP, DomainInfo.Port) {
					fmt.Println("节点异常", err)
					// 格式化消息内容，使用 Markdown 格式
					messageText = fmt.Sprintf("`节点异常`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				// 格式化消息内容，使用 Markdown 格式
				messageText = fmt.Sprintf("`节点连通性正常，正在进行A记录解析...`") // 格式化消息内容，使用 Markdown 格式
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
				_, err = services.UpdateARecord(DomainInfo.Domain, newIP)
				if err != nil {
					fmt.Println("更新域名A记录失败", err)
					// 格式化消息内容，使用 Markdown 格式
					messageText = fmt.Sprintf("`更新域名A记录失败`") // 格式化消息内容，使用 Markdown 格式
					msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					return
				}
				newDomainIp, err := repository.UpdateDomainIp(data, newIP)
				if err != nil {
					fmt.Println("更新数据库IP失败", err)
					return
				}
				ID := newDomainIp.ID
				Domain := newDomainIp.Domain
				ForwardingDomain := newDomainIp.ForwardingDomain
				IP := newDomainIp.IP
				Port := newDomainIp.Port
				ISP := newDomainIp.ISP
				Ban := newDomainIp.Ban
				// 格式化消息内容，使用 Markdown 格式
				messageText = fmt.Sprintf(
					"*检测并解析成功*✅️\nID: `%d`\n域名: `%s`\n转发域名: `%s`\nIP: `%s`\n端口: `%d`\n运营商: `%s`\nIsBan: `%t`",
					ID, Domain, ForwardingDomain, IP, Port, ISP, Ban,
				) // 格式化消息内容，使用 Markdown 格式
				fmt.Println(messageText)
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				// 创建键盘布局
				msg.ReplyMarkup = Keyboard.GenerateSubMenuKeyboard(ID, Ban)
				//发送消息
				_, err = bot.Send(msg)
			case "ban":
				// 处理封禁操作
				fmt.Println("执行封禁或启用操作, ID:", data)
				db.InitDB() //连接数据库
				DomainInfo, err := repository.GetDomainIDInfo(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				Ban := DomainInfo.Ban
				if Ban {
					newBanStatus := !DomainInfo.Ban
					_, err := repository.UpdateDomainBan(data, newBanStatus)
					if err != nil {
						fmt.Println(err)
						return
					}
					DomainInfo, err := repository.GetDomainIDInfo(data)
					if err != nil {
						fmt.Println(err)
						return
					}
					ID := DomainInfo.ID
					Domain := DomainInfo.Domain
					ForwardingDomain := DomainInfo.ForwardingDomain
					IP := DomainInfo.IP
					Port := DomainInfo.Port
					ISP := DomainInfo.ISP
					Ban := DomainInfo.Ban
					// 格式化消息内容，使用 Markdown 格式
					messageText := fmt.Sprintf(
						"*已解除封禁✅️*\nID: `%d`\n域名: `%s`\n转发域名: `%s`\nIP: `%s`\n端口: `%d`\n运营商: `%s`\nIsBan: `%t`",
						ID, Domain, ForwardingDomain, IP, Port, ISP, Ban,
					) // 格式化消息内容，使用 Markdown 格式
					fmt.Println(messageText)
					msg := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					// 创建按钮
					msg.ReplyMarkup = Keyboard.GenerateSubMenuKeyboard(ID, Ban)
					_, err = bot.Send(msg)
				} else {
					newBanStatus := !DomainInfo.Ban
					_, err := repository.UpdateDomainBan(data, newBanStatus)
					if err != nil {
						fmt.Println(err)
						return
					}
					DomainInfo, err := repository.GetDomainIDInfo(data)
					if err != nil {
						fmt.Println(err)
						return
					}
					ID := DomainInfo.ID
					Domain := DomainInfo.Domain
					ForwardingDomain := DomainInfo.ForwardingDomain
					IP := DomainInfo.IP
					Port := DomainInfo.Port
					ISP := DomainInfo.ISP
					Ban := DomainInfo.Ban
					// 格式化消息内容，使用 Markdown 格式
					messageText := fmt.Sprintf(
						"*已封禁🚫*\nID: `%d`\n域名: `%s`\n转发域名: `%s`\nIP: `%s`\n端口: `%d`\n运营商: `%s`\nIsBan: `%t`",
						ID, Domain, ForwardingDomain, IP, Port, ISP, Ban,
					) // 格式化消息内容，使用 Markdown 格式
					fmt.Println(messageText)
					msg := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					// 创建按钮
					msg.ReplyMarkup = Keyboard.GenerateSubMenuKeyboard(ID, Ban)
					_, err = bot.Send(msg)
				}
			case "back":
				// 处理退出操作
				fmt.Println("返回操作, ID:", ID)
				db.InitDB()
				DomainInfo, err := repository.GetDomainInfo()
				if err != nil {
					fmt.Println(err)
					msg := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID, // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID,
						"数据库未查询到任何域名记录❌️") // 要编辑的消息的 ID
					// 发送消息
					_, err = bot.Send(msg)
					return
				}
				keyboard := Keyboard.GenerateMainMenuKeyboard(DomainInfo) //生成内联键盘
				msg := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID, // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID,
					"查询转发信息") // 要编辑的消息的 ID
				msg.ReplyMarkup = &keyboard
				// 发送消息
				_, err = bot.Send(msg)
			case "exit":
				// 处理退出操作
				fmt.Println("退出操作, ID:", ID)
				// 删除消息
				msg := tgbotapi.NewDeleteMessage(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要删除的消息的 ID
				)
				// 发送删除消息的请求
				_, _ = bot.Send(msg)

			}
		}

		fmt.Println("当前是2级菜单")
	case 3:
		fmt.Println("当前是3级菜单")
	default:
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "无效的回调数据")
		_, _ = bot.Send(msg)
	}
}
