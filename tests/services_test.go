package tests

import (
	"fmt"
	"log"
	"telegrambot/config"
	"testing"
)

// 更新 A 记录的函数
func TestRepository(t *testing.T) {
	// 加载配置文件
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	fmt.Printf(Config.Database.Host, "11123")
}
