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
	"time"
)

var userState = make(map[int64]string)
var userMeta = make(map[int64]map[string]string)

// 新增：
var userLastPromptMessage = make(map[int64]tgbotapi.Message)

func CallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, Config *config.Config) {

	data := update.CallbackQuery.Data
	fmt.Printf(data)
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
		DomainInfoText := utils.DomainInfoText(DomainInfo, Config)
		// 格式化消息内容，使用 Markdown 格式
		messageText := fmt.Sprintf(DomainInfoText)
		fmt.Println(messageText)
		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
			update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
			messageText,                            // 新的消息文本
		)
		msg.ParseMode = "Markdown"
		// 创建按钮
		ID := DomainInfo.ID
		Ban := DomainInfo.Ban
		msg.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(ID, Ban)
		_, err = bot.Send(msg)
		fmt.Println("当前是1级菜单")
	case 2:
		if len(levels) > 1 {
			ID := levels[0]
			action := levels[1]
			switch action {
			case "weight":
				fmt.Println("设置权重, weight:", ID)
				userID := update.CallbackQuery.From.ID
				chatID := update.CallbackQuery.Message.Chat.ID
				messageID := update.CallbackQuery.Message.MessageID

				userState[userID] = "awaiting_weight_input"
				userMeta[userID] = map[string]string{"id": ID}

				editMsg := tgbotapi.NewEditMessageText(chatID, messageID, fmt.Sprintf("你正在为 ID `%s` 设置权重，请发送新的权重值（整数）", ID))
				editMsg.ParseMode = "Markdown"

				sentMsg, err := bot.Send(editMsg)
				if err == nil {
					userLastPromptMessage[userID] = sentMsg
				}
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
				keyBoard := keyboard.GenerateMainMenuKeyboard(DomainInfo) //生成内联键盘
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID, // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID,
					"记录删除成功✅️") // 要编辑的消息的 ID
				msg.ReplyMarkup = &keyBoard
				// 发送消息
				_, err = bot.Send(msg)
			case "getIp":
				fmt.Println("获取转发最新ip轮询，正在开发中....")
				// 格式化消息内容，使用 Markdown 格式
				messageText := fmt.Sprintf("`正在获取最新IP...`") // 格式化消息内容，使用 Markdown 格式
				msg := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				_, _ = bot.Send(msg)
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
				newDomainIp, err := repository.UpdateDomainIp(data, newIP)
				if err != nil {
					fmt.Println("更新数据库IP失败", err)
					return
				}
				DomainInfoText := utils.DomainInfoText(DomainInfo, Config)
				// 格式化消息内容，使用 Markdown 格式
				messageText = fmt.Sprintf(
					"*获取最新IP成功*✅\n" + DomainInfoText) // 格式化消息内容，使用 Markdown 格式
				fmt.Println(messageText)
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				// 创建按钮
				ID := newDomainIp.ID
				Ban := newDomainIp.Ban
				msg.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(ID, Ban)
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
				// 格式化消息内容，使用 Markdown 格式
				DomainInfoText := utils.DomainInfoText(DomainInfo, Config)
				messageText = fmt.Sprintf(
					"*解析成功*✅\n" + DomainInfoText) // 格式化消息内容，使用 Markdown 格式
				fmt.Println(messageText)
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				// 创建按钮
				ID := newDomainIp.ID
				Ban := newDomainIp.Ban
				msg.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(ID, Ban)
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
				BanTime := newDomainIp.BanTime + Config.BanTime.UnBanTime
				Weight := newDomainIp.Weight
				formattedTime := time.Unix(BanTime, 0).Format("2006-01-02 15:04:05")
				// 格式化消息内容，使用 Markdown 格式
				messageText = fmt.Sprintf(
					"*检测并解析成功*✅️\nID: `%d`\n域名: `%s`\n转发域名: `%s`\nIP: `%s`\n端口: `%d`\n运营商: `%s`\nIsBan: `%t`\n解禁时间: `%s`\n权重: `%d`",
					ID, Domain, ForwardingDomain, IP, Port, ISP, Ban, formattedTime, Weight) // 格式化消息内容，使用 Markdown 格式
				fmt.Println(messageText)
				msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
					messageText,                            // 新的消息文本
				)
				msg.ParseMode = "Markdown"
				// 创建键盘布局
				msg.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(ID, Ban)
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
					_, _ = repository.UpdateDomainBan(data, newBanStatus)
					_, _ = repository.UpdateDomainBanTime(data, 0)
					DomainInfo, err := repository.GetDomainIDInfo(data)
					if err != nil {
						fmt.Println(err)
						return
					}
					DomainInfoText := utils.DomainInfoText(DomainInfo, Config)
					messageText := fmt.Sprintf(
						"解除封禁✅️\n" + DomainInfoText)
					fmt.Println(messageText)
					msg := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					// 创建按钮
					ID := DomainInfo.ID
					Ban := DomainInfo.Ban
					msg.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(ID, Ban)
					_, err = bot.Send(msg)
				} else {
					newBanStatus := !DomainInfo.Ban
					_, _ = repository.UpdateDomainBan(data, newBanStatus)
					_, _ = repository.UpdateDomainBanTime(data, time.Now().AddDate(1, 0, 0).Unix())
					DomainInfo, err := repository.GetDomainIDInfo(data)
					if err != nil {
						fmt.Println(err)
						return
					}
					DomainInfoText := utils.DomainInfoText(DomainInfo, Config)
					messageText := fmt.Sprintf(
						"已封禁❌️\n" + DomainInfoText)
					fmt.Println(messageText)
					msg := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,   // 原始消息的聊天 ID
						update.CallbackQuery.Message.MessageID, // 要编辑的消息的 ID
						messageText,                            // 新的消息文本
					)
					msg.ParseMode = "Markdown"
					// 创建按钮
					ID := DomainInfo.ID
					Ban := DomainInfo.Ban
					msg.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(ID, Ban)
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
				keyBoard := keyboard.GenerateMainMenuKeyboard(DomainInfo) //生成内联键盘
				msg := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID, // 原始消息的聊天 ID
					update.CallbackQuery.Message.MessageID,
					"查询转发信息") // 要编辑的消息的 ID
				msg.ReplyMarkup = &keyBoard
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
			case "delete":
				fmt.Println("删除操作, delete:", ID)
				// 处理批量删除操作
				fmt.Println("执行是否删除操作, ID:", data)
				db.InitDB() //连接数据库
				DomainInfo, err := repository.GetDomainIDInfo(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				del := DomainInfo.Del
				if del {
					NewDelStatus := !DomainInfo.Del
					_, err := repository.UpdateDomainDelete(data, NewDelStatus)
					if err != nil {
						fmt.Println(err)
						return
					}

				} else {
					NewDelStatus := !DomainInfo.Del
					_, err := repository.UpdateDomainDelete(data, NewDelStatus)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				GetDomainInfo, err := repository.GetDomainInfo()
				if err != nil {
					fmt.Println(err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						"数据库未查询到任何域名记录❌️") // 要编辑的消息的 ID
					// 发送消息
					_, err = bot.Send(msg)
					return
				}
				// 第一步：生成消息文本和按钮
				text := "请选择删除的转发记录\n" +
					"✅️=删除\n" +
					"🚫=不删" // 或你要显示的文本
				keyboardMarkup := keyboard.GenerateMainMenuDeleteKeyboard(GetDomainInfo)

				// 第二步：编辑消息文本
				edit := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					text, // 这里是文本，不是 keyboardMarkup
				)
				edit.ParseMode = "Markdown"

				// 第三步：附加内联键盘（ReplyMarkup）
				edit.ReplyMarkup = &keyboardMarkup

				// 第四步：发送编辑请求
				_, err = bot.Send(edit)
				return
			case "confirmDel":
				fmt.Println("确认删除操作, confirmDel:", ID)
				db.InitDB() //连接数据库
				err := repository.DeleteAllMarkedDomains()
				if err != nil {
					return
				}
				GetDomainInfo, err := repository.GetDomainInfo()
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
				// 第一步：生成消息文本和按钮
				text := "✅️已删除"
				keyboardMarkup := keyboard.GenerateMainMenuDeleteKeyboard(GetDomainInfo)

				// 第二步：编辑消息文本
				edit := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					text, // 这里是文本，不是 keyboardMarkup
				)
				edit.ParseMode = "Markdown"

				// 第三步：附加内联键盘（ReplyMarkup）
				edit.ReplyMarkup = &keyboardMarkup

				// 第四步：发送编辑请求
				_, err = bot.Send(edit)
				return

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

func HandleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, Config *config.Config) {
	userID := update.Message.From.ID
	text := update.Message.Text
	chatID := update.Message.Chat.ID

	state, ok := userState[userID]
	if !ok {
		return // 无状态，忽略或正常处理
	}

	switch state {
	case "awaiting_weight_input":
		db.InitDB()
		weight, err := strconv.Atoi(text)
		if err != nil {
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "⚠️ 请输入有效的整数作为权重"))
			return
		}

		idStr := userMeta[userID]["id"]

		_, err = repository.UpdateDomainWeight(idStr, weight)
		if err != nil {
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ 权重更新失败：%v", err)))
		} else {
			DomainInfo, err := repository.GetDomainIDInfo(idStr)
			if err != nil {
				// 查询失败，发普通成功提示
				_, _ = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ 权重设置成功：ID %s → 权重 %d", idStr, weight)))
			} else {
				// 计算解禁时间
				DomainInfoText := utils.DomainInfoText(DomainInfo, Config)
				HeadText := fmt.Sprintf("✅ 权重设置成功：ID %d → 权重 %d\n", DomainInfo.ID, weight)
				promptMsg, ok := userLastPromptMessage[userID]
				if !ok {
					msg := tgbotapi.NewMessage(chatID, HeadText+DomainInfoText)
					msg.ParseMode = "Markdown" // 这里设置 Markdown 解析
					msg.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(DomainInfo.ID, DomainInfo.Ban)
					_, _ = bot.Send(msg)
				} else {
					edit := tgbotapi.NewEditMessageText(promptMsg.Chat.ID, promptMsg.MessageID, HeadText+DomainInfoText)
					edit.ParseMode = "Markdown" // 这里也设置
					edit.ReplyMarkup = keyboard.GenerateSubMenuKeyboard(DomainInfo.ID, DomainInfo.Ban)
					_, _ = bot.Send(edit)
					delete(userLastPromptMessage, userID)
				}
			}

			// 清理用户状态
			delete(userState, userID)
			delete(userMeta, userID)
		}
	}
}
