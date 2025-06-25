package services

import (
	"fmt"
	"telegrambot/config"
	"telegrambot/internal/db"
	"telegrambot/internal/db/repository"
	"time"
)

func AutoUnbanRoutine(Config *config.Config) {
	db.InitDB()
	ticker := time.NewTicker(time.Duration(Config.BanTime.CheckTime) * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		<-ticker.C

		domains, err := repository.GetDomainInfo()
		if err != nil {
			fmt.Println("获取域名失败:", err)
			continue
		}

		now := time.Now().Unix()
		fmt.Printf("%d\n", now)

		for _, domain := range domains {
			if domain.Ban && domain.BanTime > 0 {
				elapsed := now - domain.BanTime
				if elapsed >= Config.BanTime.UnBanTime { // 600秒 = 10分钟
					// 超过10分钟，自动解除封禁状态和时间
					_, errBan := repository.UpdateDomainBan(fmt.Sprintf("%d", domain.ID), false)
					_, errBanTime := repository.UpdateDomainBanTime(fmt.Sprintf("%d", domain.ID), 0)

					if errBan == nil && errBanTime == nil {
						fmt.Printf("自动解封域名 ID=%d，封禁已超过 %d 秒\n", domain.ID, elapsed)
					} else {
						fmt.Printf("自动解封失败 ID=%d, errBan=%v, errBanTime=%v\n", domain.ID, errBan, errBanTime)
					}
				}
			}
		}
	}
}
