package services

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net"
	"telegrambot/internal/db"
	"telegrambot/internal/db/repository"
	"time"
)

// CheckTCPConnectivity 对指定地址进行5次TCP连接测试，如果至少一次成功，则返回true，否则返回false
func CheckTCPConnectivity(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	success := false

	for i := 0; i < 5; i++ {
		conn, err := net.DialTimeout("tcp", address, 3*time.Second)
		if err == nil {
			success = true
			_ = conn.Close() // 关闭连接
			fmt.Printf("IP检测正常:%s\n", ip)
			break // 如果成功一次就可以退出循环
		}
		time.Sleep(1 * time.Second) // 每次尝试间隔1秒
		fmt.Printf("IP检测异常:%s---异常次数:%d\n", ip, i+1)
	}
	return success
}

// ResolveDomainToIP 解析域名并返回第一个IP地址
func ResolveDomainToIP(domain string) (string, error) {
	// 使用 net.LookupIP 解析域名
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}

	if len(ips) > 0 {
		return ips[0].String(), nil
	}

	return "", fmt.Errorf("未找到任何IP地址")
}

func ALLCheckTCPConnectivity(bot *tgbotapi.BotAPI, update tgbotapi.Update, shouldSend bool) bool {
	db.InitDB() // 连接数据库

	ALLDomain, err := repository.GetDomainInfo()
	if err != nil {
		fmt.Println("获取域名信息失败:", err)
		return false
	}

	// 消息发送/编辑处理闭包
	sendOrEditMessage := func(chatID int64, text string, messageID *int, isEdit bool) {
		if shouldSend {
			if isEdit {
				if *messageID == 0 {
					fmt.Println("错误: 尝试编辑消息，但 messageID 为 0")
					return
				}
				editMsg := tgbotapi.NewEditMessageText(chatID, *messageID, text)
				editMsg.ParseMode = "Markdown"
				if _, err := bot.Send(editMsg); err != nil {
					fmt.Printf("编辑消息失败: %v\n", err)
				}
			} else {
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = "Markdown"
				sentMsg, err := bot.Send(msg)
				if err != nil {
					fmt.Printf("发送消息失败: %v\n", err)
					return
				}
				*messageID = sentMsg.MessageID
			}
		}
	}

	// 遍历主域名
	for domainName, forwardingMap := range ALLDomain {
		var port int
		for _, details := range forwardingMap {
			if value, ok := details["Port"].(int); ok {
				port = value
			}
			break
		}

		DomainIP, err := ResolveDomainToIP(domainName)
		if err != nil {
			fmt.Printf("主域名未进行配置解析: %s\n", err)
			continue
		}

		// 主域名连通性检测
		var messageID int
		sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*开始检测域名:*`%s:%d`", domainName, port), &messageID, false)

		if isConnected := CheckTCPConnectivity(DomainIP, port); isConnected {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*节点:*`%s:%d`*正常*", domainName, port), &messageID, true)
			continue
		} else {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*节点:*`%s:%d`*无法连通*", domainName, port), &messageID, true)
		}

		// 检测子域名连通性
		for forwardingDomain, details := range forwardingMap {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*开始检测转发域名:*`%s:%d`", forwardingDomain, port), &messageID, true)

			forwardingIP, err := ResolveDomainToIP(forwardingDomain)
			if err != nil {
				fmt.Printf("转发域名解析错误: %s, 错误: %s\n", forwardingDomain, err)
				sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("转发域名解析错误:`%s`, 错误:`%s`", forwardingDomain, err), &messageID, true)
				continue
			}

			if isConnected := CheckTCPConnectivity(forwardingIP, port); isConnected {
				if _, err := UpdateARecord(domainName, forwardingIP); err != nil {
					fmt.Printf("更新域名 A 记录失败: %s\n", err)
					continue
				}

				sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("域名 A 记录更新成功:`%s`\n解析 IP:`%s`", domainName, forwardingIP), &messageID, true)

				ID := fmt.Sprintf("%v", details["ID"])
				if _, err := repository.UpdateDomainIp(ID, forwardingIP); err != nil {
					fmt.Printf("更新数据库失败: %s\n", err)
				} else {
					fmt.Printf("数据库更新成功: %s -> %s\n", forwardingDomain, forwardingIP)
				}
				break
			}
		}
	}

	fmt.Println("所有域名检测完毕")
	return true
}
