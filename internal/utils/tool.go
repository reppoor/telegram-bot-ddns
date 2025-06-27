package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"telegrambot/config"
	"telegrambot/internal/db/models"
	"time"
)

func ValidateFormat(params string) (bool, error) {
	// 使用 "#" 拆分参数
	parts := strings.Split(params, "#")

	// 检查拆分后的部分数量是否为 4
	if len(parts) != 4 {
		return false, fmt.Errorf("格式不正确请确保只有4个#号当前#号个数:%d", len(parts))
	}

	// 验证第一部分是否为有效域名格式（简单检查）
	domain := parts[0]
	if !isValidDomain(domain) {
		return false, fmt.Errorf("域名不合法，请用合法的域名格式，如www.baidu.com\n您当前传入的非法格式域名: %s", domain)
	}

	// 验证第三部分是否为整数（例如：0）
	param3 := parts[2]
	if _, err := strconv.Atoi(param3); err != nil {
		return false, fmt.Errorf("端口为非整数，请输入整数端口如7890\n您当前传入的非法格式端口: %s", param3)
	}

	// 如果所有验证都通过
	return true, nil
}

// isValidDomain 验证域名格式是否正确（支持根域名和二级域名）
func isValidDomain(domain string) bool {
	// 正则表达式检查: 根域名 或 二级域名
	// 举例: example.com, sub.example.com
	regex := `^[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z]{2,}$`
	match, err := regexp.MatchString(regex, domain)
	if err != nil {
		// 如果正则匹配出错，认为域名无效
		return false
	}
	return match
}

func DomainInfoText(domainData models.Domain, Config *config.Config) string {
	ID := domainData.ID
	Domain := domainData.Domain
	ForwardingDomain := domainData.ForwardingDomain
	IP := domainData.IP
	Port := domainData.Port
	ISP := domainData.ISP
	Ban := domainData.Ban
	BanTime := domainData.BanTime + Config.BanTime.UnBanTime
	Weight := domainData.Weight
	SortOrder := domainData.SortOrder
	formattedTime := time.Unix(BanTime, 0).Format("2006-01-02 15:04:05")

	// 状态
	banStatus := "✅ 启用中"
	if Ban {
		banStatus = "⛔ 已封禁"
	}

	messageText := fmt.Sprintf(
		"*📌 基本信息*\n"+
			"• *ID*：`%d`\n"+
			"• *排序*：`%d`\n"+
			"• *权重*：`%d`\n"+
			"——————————————\n"+
			"*🌐 域名信息*\n"+
			"• *域名*：`%s`\n"+
			"• *转发域名*：`%s`\n"+
			"• *IP地址*：`%s`\n"+
			"• *端口*：`%d`\n"+
			"• *运营商*：`%s`\n"+
			"——————————————\n"+
			"*🚦 状态信息*\n"+
			"• *当前状态*：%s\n"+
			"• *解封时间*：`%s`",
		ID, SortOrder, Weight,
		Domain, ForwardingDomain, IP, Port, ISP,
		banStatus, formattedTime,
	)

	return messageText
}
