package services

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"telegrambot/internal/db"
	"telegrambot/internal/db/repository"
	"time"
)

// 尝试通过代理进行连接，如果失败则返回错误
func tryConnectWithProxy(address string, proxyURL *url.URL) (bool, error) {
	// 创建 HTTP Transport，使用代理
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 设置连接超时
			dialer := &net.Dialer{
				Timeout: 3 * time.Second, // 设置连接超时为 3 秒
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}

	// 创建 HTTP 客户端，并使用上面的 Transport
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second, // 设置整个请求的超时
	}

	// 尝试通过 HTTP 请求检测连接
	resp, err := client.Get(fmt.Sprintf("http://%s", address))
	if err == nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return true, nil
	}

	return false, err
}

// 尝试直接连接目标地址
func tryDirectConnection(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err == nil {
		_ = conn.Close() // 关闭连接
		return true
	}
	return false
}

func CheckTCPConnectivity(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	success := false

	// 代理服务器的 URL（根据你的需求修改此代理地址）
	proxyURL, err := url.Parse("http://127.0.0.1:7890") // 替换为实际代理地址
	if err != nil {
		fmt.Println("代理地址无效:", err)
		return false
	}

	// 尝试使用代理进行连接
	for i := 0; i < 5; i++ {
		// 尝试通过代理连接
		proxySuccess, proxyErr := tryConnectWithProxy(address, proxyURL)
		if proxySuccess {
			success = true
			fmt.Printf("通过代理检测正常:%s\n", ip)
			break
		} else {
			// 如果代理连接失败，尝试直接连接
			fmt.Printf("代理连接失败，尝试直连: %v\n", proxyErr)

			directSuccess := tryDirectConnection(address)
			if directSuccess {
				success = true
				fmt.Printf("IP检测正常(直连): %s\n", ip)
				break
			} else {
				fmt.Printf("直连失败:%s---异常次数:%d\n", ip, i+1)
			}
		}

		// 每次尝试间隔1秒
		time.Sleep(1 * time.Second)
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
		// 如果使用 dnsmasq
		// cmd = exec.Command("sudo", "systemctl", "restart", "dnsmasq")
		// 如果使用 nscd
		// cmd = exec.Command("sudo", "service", "nscd", "restart")
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
