package services

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net"
	"os"
	"os/exec"
	"runtime"
	"telegrambot/config"
	"telegrambot/internal/db"
	"telegrambot/internal/db/models"
	"telegrambot/internal/db/repository"
	"time"
)

// CheckTCPConnectivity 对指定地址进行5次TCP连接测试，如果至少一次成功，则返回true，否则返回false
func CheckTCPConnectivity(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	success := false
	// 加载配置文件
	Config, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("加载配置文件失败: %v", err)
	}
	for i := 0; i < 5; i++ {
		conn, err := net.DialTimeout("tcp", address, time.Duration(Config.Check.IpCheckTime)*time.Second)
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
	db.InitDB()

	// 清除 DNS 缓存
	if err := ClearDNSCache(); err != nil {
		fmt.Println("错误:", err)
	}

	// 获取所有域名数据（按 ID 排序）
	allDomains, err := repository.GetDomainInfo()
	if err != nil {
		fmt.Println("获取域名信息失败:", err)
		return false
	}

	// 将同一个主域名的条目归类
	grouped := make(map[string][]models.Domain)
	var orderedDomainNames []string
	for _, domain := range allDomains {
		if _, ok := grouped[domain.Domain]; !ok {
			orderedDomainNames = append(orderedDomainNames, domain.Domain)
		}
		grouped[domain.Domain] = append(grouped[domain.Domain], domain)
	}

	sendOrEditMessage := func(chatID int64, text string, messageID *int, isEdit bool, forceSend bool) {
		if shouldSend || forceSend {
			if isEdit {
				if *messageID == 0 {
					msg := tgbotapi.NewMessage(chatID, text)
					msg.ParseMode = "Markdown"
					sentMsg, err := bot.Send(msg)
					if err != nil {
						fmt.Printf("发送新消息失败: %v\n", err)
						return
					}
					*messageID = sentMsg.MessageID
				} else {
					editMsg := tgbotapi.NewEditMessageText(chatID, *messageID, text)
					editMsg.ParseMode = "Markdown"
					if _, err := bot.Send(editMsg); err != nil {
						fmt.Printf("编辑消息失败: %v\n", err)
					}
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
	for _, domainName := range orderedDomainNames {
		domainEntries := grouped[domainName]
		if len(domainEntries) == 0 {
			continue
		}

		// 获取端口
		port := domainEntries[0].Port

		var messageID int

		// 检查主域名是否可连通
		DomainIP, err := ResolveDomainToIP(domainName)
		if err != nil {
			fmt.Printf("主域名未进行配置解析: %s\n", err)
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*主域名未进行配置解析记录，请先进行解析:*`%s`", domainName), &messageID, false, false)
			continue
		}

		sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*开始检测域名:*`%s:%d`", domainName, port), &messageID, false, false)
		fmt.Printf("开始检测域名:%s:%d\n", domainName, port)

		if isConnected := CheckTCPConnectivity(DomainIP, port); isConnected {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*节点:*`%s:%d`*正常*", domainName, port), &messageID, true, false)
			continue
		} else {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*节点:*`%s:%d`*无法连通*", domainName, port), &messageID, true, true)
		}

		// 转发域检测
		var forwardingDomainInfo string
		for _, item := range domainEntries {
			forwardingDomain := item.ForwardingDomain
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*开始检测转发域名:*`%s:%d`", forwardingDomain, port), &messageID, true, false)

			if item.Ban {
				msg := fmt.Sprintf("-----\n*转发域名封禁:*`%s`, *IsBan:*`%t`\n", forwardingDomain, item.Ban)
				fmt.Print(msg)
				sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)
				forwardingDomainInfo += msg
				continue
			}

			forwardingIP, err := ResolveDomainToIP(forwardingDomain)
			if err != nil {
				msg := fmt.Sprintf("-----\n*转发域名解析错误:* `%s`, 错误: %s\n", forwardingDomain, err)
				fmt.Print(msg)
				sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, false, true)
				forwardingDomainInfo += msg
				continue
			}

			fmt.Printf("开始检测转发域名:%s:%d\n", forwardingDomain, port)
			if isConnected := CheckTCPConnectivity(forwardingIP, port); isConnected {
				if _, err := UpdateARecord(domainName, forwardingIP); err != nil {
					fmt.Printf("更新域名 A 记录失败: %s\n", err)
					continue
				}

				if forwardingDomainInfo != "" {
					msg := fmt.Sprintf("*域名A记录:*`%s`\n*转发域名:*`%s`\n*解析IP:*`%s`\n*运营商:*`%s`\n=====*异常的转发域:*=====\n%s=====", domainName, forwardingDomain, forwardingIP, item.ISP, forwardingDomainInfo)
					sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)
				} else {
					msg := fmt.Sprintf("*域名A记录:*`%s`\n*转发域名:*`%s`\n*解析IP:*`%s`\n*运营商:*`%s`", domainName, forwardingDomain, forwardingIP, item.ISP)
					sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)
				}

				if _, err := repository.UpdateDomainIp(fmt.Sprintf("%d", item.ID), forwardingIP); err != nil {
					fmt.Printf("更新数据库失败: %s\n", err)
				} else {
					fmt.Printf("数据库更新成功: %s -> %s\n", forwardingDomain, forwardingIP)
				}
				break
			} else {
				msg := fmt.Sprintf("-----\n*转发域名异常:* `%s:%d`\n", forwardingDomain, port)
				fmt.Print(msg)
				forwardingDomainInfo += msg
			}
		}
	}

	fmt.Println("所有域名检测完毕")
	return true
}

// ClearDNSCache 根据操作系统清除 DNS 缓存
func ClearDNSCache() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows 系统清除 DNS 缓存
		cmd = exec.Command("ipconfig", "/flushdns")
	case "linux":
		// Linux 系统清除 DNS 缓存
		// 检查 systemd 是否存在
		cmd = exec.Command("sudo", "systemctl", "restart", "systemd-resolved")
	case "darwin": // macOS
		// macOS 系统清除 DNS 缓存
		cmd = exec.Command("sudo", "killall", "-HUP", "mDNSResponder")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	// 执行命令
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("清除 DNS 缓存失败: %v", err)
	}

	fmt.Println("DNS 缓存已成功清除")
	return nil
}
