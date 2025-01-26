package repository

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"telegrambot/internal/db"
	"telegrambot/internal/db/models"
)

func GetDomainInfo() (map[string]map[string]map[string]interface{}, error) {
	// 查询所有数据
	var domains []models.Domain
	if err := db.DB.Find(&domains).Error; err != nil {
		log.Fatalf("查询数据失败: %v", err)
		return nil, err
	}

	// 如果数据库中没有数据，返回自定义错误
	if len(domains) == 0 {
		err := fmt.Errorf("数据库中没有域名数据")
		log.Println(err)
		return nil, err
	}

	// 用于存储合并后的结果，map结构：Domain -> ForwardingDomain -> 详情
	domainMap := make(map[string]map[string]map[string]interface{})

	// 遍历所有数据并进行合并
	for _, domain := range domains {
		// 如果该 Domain 不存在，初始化它
		if _, exists := domainMap[domain.Domain]; !exists {
			domainMap[domain.Domain] = make(map[string]map[string]interface{})
		}

		// 如果该 ForwardingDomain 不存在，初始化它
		if _, exists := domainMap[domain.Domain][domain.ForwardingDomain]; !exists {
			domainMap[domain.Domain][domain.ForwardingDomain] = make(map[string]interface{})
		}

		// 将 ID, IP, Port, ISP, Ban 添加到合适的位置
		domainMap[domain.Domain][domain.ForwardingDomain]["ID"] = domain.ID
		domainMap[domain.Domain][domain.ForwardingDomain]["IP"] = domain.IP
		domainMap[domain.Domain][domain.ForwardingDomain]["Port"] = domain.Port
		domainMap[domain.Domain][domain.ForwardingDomain]["ISP"] = domain.ISP
		domainMap[domain.Domain][domain.ForwardingDomain]["Ban"] = domain.Ban
	}

	return domainMap, nil
}

func GetDomainIDInfo(ID string) (domainInfo models.Domain, err error) {
	var domain models.Domain
	// 初始化默认值
	var numericID = ID

	// 检查并提取 ID 的数字部分（如果包含 "-"）
	if strings.Contains(ID, "-") {
		idParts := strings.Split(ID, "-")
		if len(idParts) > 0 {
			numericID = idParts[0] // 提取 "-" 前的部分
		}
	}
	// 将字符串ID转换为uint类型
	uintID, err := strconv.ParseUint(numericID, 10, 32)
	if err != nil {
		fmt.Printf("无效的ID格式: %v\n", err)
		return domain, err
	}

	// 根据ID查询
	result := db.DB.First(&domain, uint(uintID))
	if result.Error != nil {
		if result.RowsAffected == 0 {
			fmt.Println("未找到记录")
		} else {
			fmt.Printf("查询错误: %v\n", result.Error)
		}
		return domain, err
	}

	// 输出查询结果
	//fmt.Printf("查询结果: %+v\n", domain)
	return domain, nil
}

func UpdateDomainIp(ID string, newIP string) (models.Domain, error) {
	// 初始化默认值
	var numericID = ID

	// 检查并提取 ID 的数字部分（如果包含 "-"）
	if strings.Contains(ID, "-") {
		idParts := strings.Split(ID, "-")
		if len(idParts) > 0 {
			numericID = idParts[0] // 提取 "-" 前的部分
		}
	}
	// 转换字符串 ID 为 uint 类型
	uintID, err := strconv.ParseUint(numericID, 10, 32) // 将字符串ID转换为uint类型
	if err != nil {
		fmt.Printf("无效的ID格式: %v\n", err)
		return models.Domain{}, err
	}

	// 查找目标域名记录
	var domain models.Domain
	result := db.DB.First(&domain, uint(uintID))
	if result.Error != nil {
		fmt.Printf("查询失败: %v\n", result.Error)
		return models.Domain{}, result.Error
	}

	// 更新IP地址
	domain.IP = newIP
	updateResult := db.DB.Save(&domain)
	if updateResult.Error != nil {
		fmt.Printf("更新失败: %v\n", updateResult.Error)
		return models.Domain{}, updateResult.Error
	}

	// 返回更新后的记录
	return domain, nil
}

func DeleteDomainByID(ID string) (models.Domain, error) {
	var domain models.Domain
	// 初始化默认值
	var numericID = ID

	// 检查并提取 ID 的数字部分（如果包含 "-"）
	if strings.Contains(ID, "-") {
		idParts := strings.Split(ID, "-")
		if len(idParts) > 0 {
			numericID = idParts[0] // 提取 "-" 前的部分
		}
	}
	// 将字符串ID转换为uint类型
	uintID, err := strconv.ParseUint(numericID, 10, 32)
	if err != nil {
		fmt.Printf("无效的ID格式: %v\n", err)
		return domain, err
	}

	// 根据ID查询
	result := db.DB.First(&domain, uint(uintID))
	if result.Error != nil {
		if result.RowsAffected == 0 {
			fmt.Println("未找到记录")
		} else {
			fmt.Printf("查询错误: %v\n", result.Error)
		}
		return domain, err
	}

	// 删除记录
	deleteResult := db.DB.Delete(&domain)
	if deleteResult.Error != nil {
		fmt.Printf("删除错误: %v\n", deleteResult.Error)
		return domain, deleteResult.Error
	}

	// 输出删除结果
	fmt.Println("记录已删除")
	return domain, nil
}

func UpdateDomainBan(ID string, Ban bool) (models.Domain, error) {
	// 初始化默认值
	var numericID = ID

	// 检查并提取 ID 的数字部分（如果包含 "-"）
	if strings.Contains(ID, "-") {
		idParts := strings.Split(ID, "-")
		if len(idParts) > 0 {
			numericID = idParts[0] // 提取 "-" 前的部分
		}
	}
	// 转换字符串 ID 为 uint 类型
	uintID, err := strconv.ParseUint(numericID, 10, 32) // 将字符串ID转换为uint类型
	if err != nil {
		fmt.Printf("无效的ID格式: %v\n", err)
		return models.Domain{}, err
	}

	// 查找目标域名记录
	var domain models.Domain
	result := db.DB.First(&domain, uint(uintID))
	if result.Error != nil {
		fmt.Printf("查询失败: %v\n", result.Error)
		return models.Domain{}, result.Error
	}

	// 更新IP地址
	domain.Ban = Ban
	updateResult := db.DB.Save(&domain)
	if updateResult.Error != nil {
		fmt.Printf("更新失败: %v\n", updateResult.Error)
		return models.Domain{}, updateResult.Error
	}

	// 返回更新后的记录
	return domain, nil
}

func InsertDomainInfo(Domain string, ForwardingDomain string, Port int, ISP string) (models.Domain, error) {
	// 先查询数据库，检查是否存在相同的 ForwardingDomain 和 Port 组合
	var existingDomain models.Domain
	if err := db.DB.Where("Domain = ? AND forwarding_domain = ? AND port = ?", Domain, ForwardingDomain, Port).First(&existingDomain).Error; err == nil {
		// 如果存在记录，则返回错误，表示该记录已经存在
		return models.Domain{}, fmt.Errorf("已存在域名 '%s' 转发域名 '%s' and 端口 '%d", Domain, ForwardingDomain, Port)
	} else if err != gorm.ErrRecordNotFound {
		// 如果发生了其他错误（非记录未找到），则返回错误
		return models.Domain{}, fmt.Errorf("非记录未找到info: %v", err)
	}
	DomainInfo := models.Domain{
		Domain:           Domain,
		ForwardingDomain: ForwardingDomain,
		Port:             Port,
		IP:               "0",
		ISP:              ISP,
		Ban:              false,
	}
	// 将新的记录插入数据库
	if err := db.DB.Create(&DomainInfo).Error; err != nil {
		return models.Domain{}, fmt.Errorf("插入数据库失败info: %v", err)
	}

	return DomainInfo, nil
}

func GetDomainInfoByIp(Domain string, ip string) (domainInfo models.Domain, err error) {
	domain := models.Domain{
		Domain:           Domain,
		ForwardingDomain: "",
		IP:               ip,
		Port:             0,
		ISP:              "",
	}

	// 根据域名和IP查询记录
	result := db.DB.Where("domain = ? AND ip = ?", Domain, ip).First(&domain)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			fmt.Println("未找到记录")
		} else {
			fmt.Printf("查询错误: %v\n", result.Error)
		}
		return domain, fmt.Errorf("查询错误: %v", err)
	}

	// 输出查询结果
	//fmt.Printf("查询结果: %+v\n", domain)
	return domain, nil
}

func GetALLDomain() ([]models.Domain, error) {
	// 查询所有数据
	var domains []models.Domain
	if err := db.DB.Find(&domains).Error; err != nil {
		// 记录日志，但不终止程序
		log.Printf("查询数据失败: %v", err)
		return nil, err
	}

	// 检查是否查询到结果
	if len(domains) == 0 {
		log.Println("没有找到任何域名数据")
		return nil, nil
	}

	return domains, nil
}
