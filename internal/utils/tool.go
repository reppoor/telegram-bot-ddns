package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// isValidDomain 验证域名格式是否正确（只允许出现两个点）
func isValidDomain(domain string) bool {
	// 正则表达式检查: 子域名.主域名.顶级域名
	// 举例: www.baidu.com, sub.example.org
	regex := `^[a-zA-Z0-9-]+\.[a-zA-Z0-9-]+\.[a-zA-Z]{2,}$`
	match, err := regexp.MatchString(regex, domain)
	if err != nil {
		// 如果正则匹配出错，认为域名无效
		return false
	}
	return match
}
