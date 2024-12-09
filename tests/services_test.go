package tests

import (
	"telegrambot/internal/services"
	"testing"
)

// 更新 A 记录的函数
func TestRepository(t *testing.T) {
	_ = services.ALLCheckTCPConnectivity()

}
