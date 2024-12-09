package services

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"log"
	"strings"
	"telegrambot/config"
)

// UpdateARecord 更新 A 记录的函数
func UpdateARecord(fullDomain, ip string) (string string, err error) {
	// 创建 Cloudflare 客户端
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Panic(err)
	}
	key := Config.Cloudflare.Key     // 替换为你的 Cloudflare API 密钥
	email := Config.Cloudflare.Email // 替换为你的 Cloudflare 电子邮件地址
	client, err := cloudflare.New(key, email)
	if err != nil {
		return "", fmt.Errorf("创建 Cloudflare 客户端失败: %v", err)
	}
	ctx := context.Background()
	fmt.Println("创建 Cloudflare 客户端成功")
	// 分割域名为子域名和主域名
	parts := strings.Split(fullDomain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("无效的域名: %s", fullDomain)
	}

	domain := parts[len(parts)-2] + "." + parts[len(parts)-1] // 获取主域名
	//subdomain := fullDomain[:len(fullDomain)-len(domain)-1]   // 获取子域名
	rc := &cloudflare.ResourceContainer{
		Level:      "",
		Identifier: "",
		Type:       "",
	}
	paramsListDNSRecordsParams := cloudflare.ListDNSRecordsParams{
		Type:       "",
		Name:       "",
		Content:    "",
		Proxied:    nil,
		Comment:    "",
		Tags:       nil,
		TagMatch:   "",
		Order:      "",
		Direction:  "",
		Match:      "",
		Priority:   nil,
		ResultInfo: cloudflare.ResultInfo{},
	}
	params := cloudflare.UpdateDNSRecordParams{
		Type:     "",
		Name:     "",
		Content:  ip,
		Data:     nil,
		ID:       "",
		Priority: nil,
		TTL:      60,
		Proxied:  nil,
		Comment:  nil,
		Tags:     nil,
	}
	// 获取域名的 Zone ID
	zones, err := client.ListZones(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Zone 列表失败: %v", err)
	}
	// 在区域列表中查找匹配的主域名并返回其 Zone ID
	for _, zone := range zones {
		//fmt.Println(zone.Name)
		if zone.Name == domain {
			rc.Identifier = zone.ID
			DNSRecord, _, err := client.ListDNSRecords(ctx, rc, paramsListDNSRecordsParams)
			if err != nil {
				return "", err
			}
			for _, zone2 := range DNSRecord {
				if zone2.Name == fullDomain {
					params.Type = "A"
					params.Name = zone2.Name
					params.ID = zone2.ID
					break
				}
			}
		}
	}
	_, err = client.UpdateDNSRecord(ctx, rc, params)
	if err != nil {
		fmt.Println("更新失败", err)
		return "", err
	}
	// 如果找不到匹配的 Zone ID，返回错误
	//return fmt.Errorf("未找到子域名 %s 对应的 Zone ID", fullDomain)
	return "域名解析成功", nil
}
