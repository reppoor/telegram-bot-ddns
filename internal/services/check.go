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
		conn, err := net.DialTimeout("tcp", address, Config.Check.IpCheckTime*time.Second)
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
	// 调用清除 DNS 缓存的函数
	if err := ClearDNSCache(); err != nil {
		fmt.Println("错误:", err)
	}
	ALLDomain, err := repository.GetDomainInfo()
	if err != nil {
		fmt.Println("获取域名信息失败:", err)
		return false
	}

	sendOrEditMessage := func(chatID int64, text string, messageID *int, isEdit bool, forceSend bool) {
		// 如果是静默模式但 forceSend 为 true，强制发送消息
		if shouldSend || forceSend {
			if isEdit {
				if *messageID == 0 {
					// 如果 messageID 为 0，尝试发送新消息
					fmt.Println("错误: 尝试编辑消息，但 messageID 为 0，发送新消息")
					msg := tgbotapi.NewMessage(chatID, text)
					msg.ParseMode = "Markdown"
					sentMsg, err := bot.Send(msg)
					if err != nil {
						fmt.Printf("发送新消息失败: %v\n", err)
						return
					}
					*messageID = sentMsg.MessageID // 更新 messageID 为发送的消息ID
				} else {
					// 编辑已有消息
					editMsg := tgbotapi.NewEditMessageText(chatID, *messageID, text)
					editMsg.ParseMode = "Markdown"
					if _, err := bot.Send(editMsg); err != nil {
						fmt.Printf("编辑消息失败: %v\n", err)
					}
				}
			} else {
				// 发送新消息
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = "Markdown"
				sentMsg, err := bot.Send(msg)
				if err != nil {
					fmt.Printf("发送消息失败: %v\n", err)
					return
				}
				*messageID = sentMsg.MessageID // 更新 messageID 为发送的消息ID
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
		sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*开始检测域名:*`%s:%d`", domainName, port), &messageID, false, false)
		fmt.Printf("开始检测域名:%s:%d\n", domainName, port)
		if isConnected := CheckTCPConnectivity(DomainIP, port); isConnected {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*节点:*`%s:%d`*正常*", domainName, port), &messageID, true, false)
			continue
		} else {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*节点:*`%s:%d`*无法连通*", domainName, port), &messageID, true, true)
		}

		// 检测子域名连通性
		for forwardingDomain, details := range forwardingMap {
			sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*开始检测转发域名:*`%s:%d`", forwardingDomain, port), &messageID, true, false)
			ban, _ := details["Ban"].(bool)
			if ban {
				sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*转发域名封禁:*`%s`, *IsBan:*`%t`", forwardingDomain, ban), &messageID, true, true)
				continue
			}

			forwardingIP, err := ResolveDomainToIP(forwardingDomain)
			if err != nil {
				fmt.Printf("转发域名解析错误: %s, 错误: %s\n", forwardingDomain, err)
				sendOrEditMessage(update.Message.Chat.ID, fmt.Sprintf("*转发域名解析错误:*`%s`, *错误:*`%s`", forwardingDomain, err), &messageID, true, true)
				continue
			}
			fmt.Printf("开始检测转发域名:%s:%d\n", forwardingDomain, port)
			if isConnected := CheckTCPConnectivity(forwardingIP, port); isConnected {
				if _, err := UpdateARecord(domainName, forwardingIP); err != nil {
					fmt.Printf("更新域名 A 记录失败: %s\n", err)
					continue
				}
				msg := fmt.Sprintf("*域名A记录:*`%s`\n*转发域名:*`%s\n`*解析IP:*`%s`\n*运营商:*`%s`", domainName, forwardingDomain, forwardingIP, details["ISP"])
				sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)

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
