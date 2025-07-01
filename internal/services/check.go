package services

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
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
	db.InitDB()
	if err := ClearDNSCache(); err != nil {
		fmt.Println("错误:", err)
	}
	allDomains, err := repository.GetDomainInfo()
	if err != nil {
		fmt.Println("获取域名信息失败:", err)
		return false
	}
	grouped := make(map[string][]models.Domain)
	var orderedDomainNames []string
	for _, domain := range allDomains {
		if _, ok := grouped[domain.Domain]; !ok {
			orderedDomainNames = append(orderedDomainNames, domain.Domain)
		}
		grouped[domain.Domain] = append(grouped[domain.Domain], domain)
	}
	for _, domainName := range orderedDomainNames {
		domainEntries := grouped[domainName]
		if len(domainEntries) == 0 {
			continue
		}
		sort.Slice(domainEntries, func(i, j int) bool {
			return domainEntries[i].Weight > domainEntries[j].Weight
		})
		fmt.Printf("主域名 %s 的转发域排序如下：\n", domainName)
		for _, item := range domainEntries {
			fmt.Printf("  转发域名: %s, 权重: %d, 封禁: %t\n", item.ForwardingDomain, item.Weight, item.Ban)
		}
		port := domainEntries[0].Port
		var messageID int
		DomainIP, err := ResolveDomainToIP(domainName)
		if err != nil {
			sendOrEditMessage(update.Message.Chat.ID,
				fmt.Sprintf("*❗️ 主域名未进行配置解析，请先配置:* `%s`", domainName), &messageID, false, false)
			continue
		}
		sendOrEditMessage(update.Message.Chat.ID,
			fmt.Sprintf("*📡 开始检测主域名:* `%s:%d`", domainName, port), &messageID, false, false)
		if isConnected := CheckTCPConnectivity(DomainIP, port); isConnected {
			sendOrEditMessage(update.Message.Chat.ID,
				fmt.Sprintf("*✅ 节点连通:* `%s:%d`", domainName, port), &messageID, true, false)
			continue
		} else {
			sendOrEditMessage(update.Message.Chat.ID,
				fmt.Sprintf("*❌ 主节点不可达:* `%s:%d`\n🔄 尝试转发节点...", domainName, port), &messageID, true, true)
		}
		var forwardingDomainInfo string
		for _, item := range domainEntries {
			forwardingDomain := item.ForwardingDomain
			ISP := item.ISP
			RecordType := item.RecordType
			sendOrEditMessage(update.Message.Chat.ID,
				fmt.Sprintf("*🔍 检测转发域名:* `%s:%d` (权重: `%d`)", forwardingDomain, port, item.Weight), &messageID, true, false)
			if item.Ban {
				msg := fmt.Sprintf("❌ *转发域名已封禁:* `%s`\n封禁状态: `true`\n权重: `%d`\n", forwardingDomain, item.Weight)
				sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)
				forwardingDomainInfo += msg
				continue
			}
			forwardingIP, err := ResolveDomainToIP(forwardingDomain)
			if err != nil {
				msg := fmt.Sprintf("❗️ *转发域名解析失败:* `%s`\n错误信息: `%s`\n➡️ 已封禁该节点\n", forwardingDomain, err.Error())
				sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, false, true)
				_, _ = repository.UpdateDomainBan(strconv.Itoa(int(item.ID)), true)
				_, _ = repository.UpdateDomainBanTime(strconv.Itoa(int(item.ID)), time.Now().AddDate(1, 0, 0).Unix())
				forwardingDomainInfo += msg
				continue
			}
			if isConnected := CheckTCPConnectivity(forwardingIP, port); isConnected {
				if RecordType {
					if _, err := UpdateARecord(domainName, forwardingIP, ISP); err != nil {
						fmt.Printf("更新域名 A 记录失败: %s\n", err)
						continue
					}
					msg := fmt.Sprintf("*✅ 成功切换 A 记录*\n🌐 *主域名:* `%s`\n🔀 *转发域名:* `%s`\n📥 *解析IP:* `%s`\n🏢 *运营商:* `%s`\n%s",
						domainName, forwardingDomain, forwardingIP, item.ISP, forwardingDomainInfo)
					sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)
					if _, err := repository.UpdateDomainIp(fmt.Sprintf("%d", item.ID), forwardingIP); err != nil {
						fmt.Printf("更新数据库失败: %s\n", err)
					} else {
						fmt.Printf("数据库更新成功: %s -> %s\n", forwardingDomain, forwardingIP)
					}
					break
				} else {
					if _, err := UpdateCNAMERecord(domainName, forwardingDomain, ISP); err != nil {
						fmt.Printf("更新域名 CNAME 记录失败: %s\n", err)
						continue
					}
					msg := fmt.Sprintf("*✅ 成功切换 CNAME 记录*\n🌐 *主域名:* `%s`\n🔀 *转发域名:* `%s`\n📥 *解析IP:* `%s`\n🏢 *运营商:* `%s`\n%s",
						domainName, forwardingDomain, forwardingIP, item.ISP, forwardingDomainInfo)
					sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)
					if _, err := repository.UpdateDomainIp(fmt.Sprintf("%d", item.ID), forwardingIP); err != nil {
						fmt.Printf("更新数据库失败: %s\n", err)
					} else {
						fmt.Printf("数据库更新成功: %s -> %s\n", forwardingDomain, forwardingIP)
					}
					break
				}
			} else {
				msg := fmt.Sprintf("❌ *转发域名不可达:* `%s:%d`\n(权重: `%d`)\n➡️ 已封禁\n", forwardingDomain, port, item.Weight)
				sendOrEditMessage(update.Message.Chat.ID, msg, &messageID, true, true)
				_, _ = repository.UpdateDomainBan(strconv.Itoa(int(item.ID)), true)
				_, _ = repository.UpdateDomainBanTime(strconv.Itoa(int(item.ID)), time.Now().Unix())
				forwardingDomainInfo += msg
			}
		}
		sendOrEditMessage(update.Message.Chat.ID, "*✅ 所有域名检测完成*", &messageID, false, true)
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
