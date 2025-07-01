package services

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"log"
	"strings"
	"telegrambot/config"
	"telegrambot/internal/db/models"
	"telegrambot/internal/db/repository"
)

func StringPtr(s string) *string {
	return &s
}

// UpdateARecord 更新 A 记录的函数，目前支持根域名与二级域名
func UpdateARecord(fullDomain, ip string, RemarkInfo string) (string, error) {
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

	// 分割域名为子域名和主域名
	parts := strings.Split(fullDomain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("无效的域名: %s", fullDomain)
	}

	// 判断是否为根域名（example.com）
	var subdomain string
	if len(parts) == 2 {
		subdomain = "" // 根域名没有子域名
	} else {
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

	// 获取域名的 Zone ID
	zones, err := client.ListZones(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Zone 列表失败: %v", err)
	}

	var zoneID string
	for _, zone := range zones {
		if zone.Name == domain {
			zoneID = zone.ID
			break
		}
	}

	if zoneID == "" {
		return "", fmt.Errorf("未找到主域名 %s 对应的 Zone ID", domain)
	}

	// 列出 DNS 记录
	rc := &cloudflare.ResourceContainer{
		Level:      "zone",
		Identifier: zoneID,
	}
	DNSRecords, _, err := client.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return "", fmt.Errorf("列出 DNS 记录失败: %v", err)
	}

	// 输出所有 DNS 记录，帮助调试
	for _, record := range DNSRecords {
		fmt.Printf("记录：ID=%s, 名称=%s, 类型=%s, 内容=%s\n", record.ID, record.Name, record.Type, record.Content)
	}

	var recordID string
	for _, record := range DNSRecords {
		// 如果是根域名或者完整匹配子域名
		if (subdomain == "" && record.Name == domain) || record.Name == fullDomain {
			recordID = record.ID
			fmt.Printf("匹配的记录ID: %s\n", recordID) // 输出匹配的记录ID
			break
		}
	}

	// 如果没有找到记录
	if recordID == "" {
		return "", fmt.Errorf("未找到匹配的 DNS 记录：%s", fullDomain)
	}
	var commentPtr *string
	if RemarkInfo == "nil" {
		commentPtr = nil
	} else {
		commentPtr = StringPtr(RemarkInfo)
	}
	// 更新 DNS 记录
	params := cloudflare.UpdateDNSRecordParams{
		Type:    "A",
		Name:    fullDomain,
		Content: ip,
		TTL:     60,
		ID:      recordID,   // 确保 ID 被正确设置
		Comment: commentPtr, //
	}
	_, err = client.UpdateDNSRecord(ctx, rc, params)
	if err != nil {
		return "", fmt.Errorf("更新 DNS 记录失败: %v", err)
	}

	// 返回成功信息
	return "域名解析成功", nil
}
func UpdateCNAMERecord(fullDomain, CNAMEDomain string, RemarkInfo string) (string, error) {
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

	// 分割域名为子域名和主域名
	parts := strings.Split(fullDomain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("无效的域名: %s", fullDomain)
	}

	// 判断是否为根域名（example.com）
	var subdomain string
	if len(parts) == 2 {
		subdomain = "" // 根域名没有子域名
	} else {
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

	// 获取域名的 Zone ID
	zones, err := client.ListZones(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Zone 列表失败: %v", err)
	}

	var zoneID string
	for _, zone := range zones {
		if zone.Name == domain {
			zoneID = zone.ID
			break
		}
	}

	if zoneID == "" {
		return "", fmt.Errorf("未找到主域名 %s 对应的 Zone ID", domain)
	}

	// 列出 DNS 记录
	rc := &cloudflare.ResourceContainer{
		Level:      "zone",
		Identifier: zoneID,
	}
	DNSRecords, _, err := client.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return "", fmt.Errorf("列出 DNS 记录失败: %v", err)
	}

	// 输出所有 DNS 记录，帮助调试
	for _, record := range DNSRecords {
		fmt.Printf("记录：ID=%s, 名称=%s, 类型=%s, 内容=%s\n", record.ID, record.Name, record.Type, record.Content)
	}

	var recordID string
	for _, record := range DNSRecords {
		// 如果是根域名或者完整匹配子域名
		if (subdomain == "" && record.Name == domain) || record.Name == fullDomain {
			recordID = record.ID
			fmt.Printf("匹配的记录ID: %s\n", recordID) // 输出匹配的记录ID
			break
		}
	}

	// 如果没有找到记录
	if recordID == "" {
		return "", fmt.Errorf("未找到匹配的 DNS 记录：%s", fullDomain)
	}

	var commentPtr *string
	if RemarkInfo == "nil" {
		commentPtr = nil
	} else {
		commentPtr = StringPtr(RemarkInfo)
	}
	// 更新 DNS 记录
	params := cloudflare.UpdateDNSRecordParams{
		Type:    "CNAME",
		Name:    fullDomain,
		Content: CNAMEDomain,
		TTL:     60,
		ID:      recordID,   // 确保 ID 被正确设置
		Comment: commentPtr, //
	}
	_, err = client.UpdateDNSRecord(ctx, rc, params)
	if err != nil {
		return "", fmt.Errorf("更新 DNS 记录失败: %v", err)
	}

	// 返回成功信息
	return "域名解析成功", nil
}

func GetDomainInfo(fullDomain string) (models.Domain, error) {
	var Domain models.Domain
	// 创建 Cloudflare 客户端
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Panic(err)
	}
	key := Config.Cloudflare.Key     // 替换为你的 Cloudflare API 密钥
	email := Config.Cloudflare.Email // 替换为你的 Cloudflare 电子邮件地址
	client, err := cloudflare.New(key, email)
	if err != nil {
		return Domain, fmt.Errorf("创建 Cloudflare 客户端失败: %v", err)
	}
	ctx := context.Background()

	// 分割域名为子域名和主域名
	parts := strings.Split(fullDomain, ".")
	if len(parts) < 2 {
		return Domain, fmt.Errorf("无效的域名: %s", fullDomain)
	}

	// 判断是否为根域名（example.com）
	var subdomain string
	if len(parts) == 2 {
		subdomain = "" // 根域名没有子域名
	} else {
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

	// 获取域名的 Zone ID
	zones, err := client.ListZones(ctx)
	if err != nil {
		return Domain, fmt.Errorf("获取 Zone 列表失败: %v", err)
	}

	var zoneID string
	for _, zone := range zones {
		if zone.Name == domain {
			zoneID = zone.ID
			break
		}
	}

	if zoneID == "" {
		return Domain, fmt.Errorf("未找到主域名 %s 对应的 Zone ID", domain)
	}

	// 列出 DNS 记录
	rc := &cloudflare.ResourceContainer{
		Level:      "zone",
		Identifier: zoneID,
	}
	DNSRecords, _, err := client.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return Domain, fmt.Errorf("列出 DNS 记录失败: %v", err)
	}

	// 输出所有 DNS 记录，帮助调试
	//for _, record := range DNSRecords {
	//fmt.Printf("记录：ID=%s, 名称=%s, 类型=%s, 解析内容=%s\n", record.ID, record.Name, record.Type, record.Content)
	//}

	var recordID string
	var recordIP string
	for _, record := range DNSRecords {
		// 如果是根域名或者完整匹配子域名
		if (subdomain == "" && record.Name == domain) || record.Name == fullDomain {
			recordID = record.ID
			recordIP = record.Content
			fmt.Printf("匹配的记录ID: %s\n", recordID) // 输出匹配的记录ID
			break
		}
	}
	// 如果没有找到记录
	if recordID == "" {
		return Domain, fmt.Errorf("未找到匹配的 DNS 记录：%s", fullDomain)
	}
	// 进行数据库查询根据IP查询
	fmt.Printf(recordIP)
	DomainInfo, err := repository.GetDomainInfoByIp(fullDomain, recordIP)
	if err != nil {
		return Domain, err
	}
	return DomainInfo, nil
}

func GetCloudflareDomainInfo(fullDomain string) (models.Domain, error) {
	var Domain models.Domain
	// 创建 Cloudflare 客户端
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Panic(err)
	}
	key := Config.Cloudflare.Key     // 替换为你的 Cloudflare API 密钥
	email := Config.Cloudflare.Email // 替换为你的 Cloudflare 电子邮件地址
	client, err := cloudflare.New(key, email)
	if err != nil {
		return Domain, fmt.Errorf("创建 Cloudflare 客户端失败: %v", err)
	}
	ctx := context.Background()

	// 分割域名为子域名和主域名
	parts := strings.Split(fullDomain, ".")
	if len(parts) < 2 {
		return Domain, fmt.Errorf("无效的域名: %s", fullDomain)
	}

	// 判断是否为根域名（example.com）
	var subdomain string
	if len(parts) == 2 {
		subdomain = "" // 根域名没有子域名
	} else {
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

	// 获取域名的 Zone ID
	zones, err := client.ListZones(ctx)
	if err != nil {
		return Domain, fmt.Errorf("获取 Zone 列表失败: %v", err)
	}

	var zoneID string
	for _, zone := range zones {
		if zone.Name == domain {
			zoneID = zone.ID
			break
		}
	}

	if zoneID == "" {
		return Domain, fmt.Errorf("未找到主域名 %s 对应的 Zone ID", domain)
	}

	// 列出 DNS 记录
	rc := &cloudflare.ResourceContainer{
		Level:      "zone",
		Identifier: zoneID,
	}
	DNSRecords, _, err := client.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return Domain, fmt.Errorf("列出 DNS 记录失败: %v", err)
	}

	// 输出所有 DNS 记录，帮助调试
	//for _, record := range DNSRecords {
	//fmt.Printf("记录：ID=%s, 名称=%s, 类型=%s, 解析内容=%s\n", record.ID, record.Name, record.Type, record.Content)
	//}

	var recordID string
	var recordAddress string
	var recordComment string
	var recordType string
	for _, record := range DNSRecords {
		// 如果是根域名或者完整匹配子域名
		if (subdomain == "" && record.Name == domain) || record.Name == fullDomain {
			recordID = record.ID
			recordAddress = record.Content
			recordComment = record.Comment
			recordType = record.Type
			fmt.Printf("匹配的记录ID: %s\n", recordID) // 输出匹配的记录ID
			break
		}
	}
	// 如果没有找到记录
	if recordID == "" {
		return Domain, fmt.Errorf("未找到匹配的 DNS 记录：%s", fullDomain)
	}
	ip, _ := ResolveDomainToIP(recordAddress)
	var RecordType bool
	if recordType == "A" {
		RecordType = true
	} else {
		RecordType = false
	}
	DomainInfo := models.Domain{
		ID:               0,
		Domain:           fullDomain,
		ForwardingDomain: recordAddress,
		IP:               ip,
		Port:             0,
		ISP:              recordComment,
		Ban:              false,
		Del:              false,
		Weight:           0,
		BanTime:          0,
		SortOrder:        0,
		RecordType:       RecordType,
	}
	return DomainInfo, nil
}
