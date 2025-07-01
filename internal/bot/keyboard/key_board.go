package keyboard

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegrambot/internal/db/models"
)

// Button 单个按钮结构体
type Button struct {
	Text         string
	CallbackData string
}

// InlineKeyboard 内联键盘结构体，包含二维按钮数组
type InlineKeyboard struct {
	Buttons [][]Button
}

func createInlineKeyboard(keyboard InlineKeyboard, class int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// 使用 keyboard.Buttons 替代 buttons
	for _, buttonRow := range keyboard.Buttons {
		var row []tgbotapi.InlineKeyboardButton
		for _, button := range buttonRow {
			// 使用回调数据作为按钮的回调值
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(button.Text, button.CallbackData))
		}
		rows = append(rows, row)
	}
	if class == 1 {
		// 创建退出按钮并添加到最后一行
		exitButton := tgbotapi.NewInlineKeyboardButtonData("退出🔚", "1-exit")
		rows = append(rows, []tgbotapi.InlineKeyboardButton{exitButton})
	} else if class == 2 {
		confirmButton := tgbotapi.NewInlineKeyboardButtonData("确认删除✅", "1-confirmDel")
		exitButton := tgbotapi.NewInlineKeyboardButtonData("退出🔚", "1-exit")

		// 将确认和退出按钮放在同一行
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(confirmButton, exitButton))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func GenerateMainMenuKeyboard(domains []models.Domain) tgbotapi.InlineKeyboardMarkup {
	var keyboard InlineKeyboard

	for _, domain := range domains {
		// 设置状态图标
		statusEmoji := "✅" // 启用中
		if domain.Ban {
			statusEmoji = "⛔" // 已封禁
		}

		// 设置按钮文本：域名 + 端口 + 状态
		buttonText := fmt.Sprintf("%s:%d [%s] %s", domain.Domain, domain.Port, domain.ForwardingDomain, statusEmoji)
		callbackData := fmt.Sprintf("%d", domain.ID)

		button := Button{
			Text:         buttonText,
			CallbackData: callbackData,
		}

		// 每个按钮单独一行
		keyboard.Buttons = append(keyboard.Buttons, []Button{button})
	}

	return createInlineKeyboard(keyboard, 1)
}

func GenerateMainMenuDeleteKeyboard(domains []models.Domain) tgbotapi.InlineKeyboardMarkup {
	var keyboard InlineKeyboard

	for _, domain := range domains {
		// 删除状态图标：✅ 已标记删除，🟡 未标记
		delEmoji := "❌"
		if domain.Del {
			delEmoji = "✅"
		}

		// 按钮文本：图标 + 域名:端口 + 转发域名（用圆括号包起来）
		buttonText := fmt.Sprintf("%s  %s:%d  （%s）", delEmoji, domain.Domain, domain.Port, domain.ForwardingDomain)
		callbackData := fmt.Sprintf("%d-delete", domain.ID)

		button := Button{
			Text:         buttonText,
			CallbackData: callbackData,
		}

		// 单列，每个按钮一行
		keyboard.Buttons = append(keyboard.Buttons, []Button{button})
	}

	return createInlineKeyboard(keyboard, 2) // 1列单排
}

func GenerateSubMenuKeyboard(ID uint, Ban bool) *tgbotapi.InlineKeyboardMarkup {
	// 封禁状态按钮文本
	banText := "✅ 启用中"
	if Ban {
		banText = "⛔ 已封禁"
	}

	// 第一行：状态相关按钮，拆成两行，减少拥挤
	row1 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(banText, fmt.Sprintf("%d-ban", ID)),
		tgbotapi.NewInlineKeyboardButtonData("✏️记录变更", fmt.Sprintf("%d-record", ID)),
	)
	row2 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⚙️ 权重", fmt.Sprintf("%d-weight", ID)),
		tgbotapi.NewInlineKeyboardButtonData("↕️ 排序", fmt.Sprintf("%d-sort", ID)),
	)

	// 第二组：解析相关操作
	row3 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🌐 获取IP", fmt.Sprintf("%d-getIp", ID)),
		tgbotapi.NewInlineKeyboardButtonData("📡 解析记录", fmt.Sprintf("%d-parse", ID)),
	)
	row4 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 检测解析", fmt.Sprintf("%d-checkAndParse", ID)),
	)

	// 第三组：删除和退出操作
	row5 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ 删除记录", fmt.Sprintf("%d-del", ID)),
	)
	row6 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 返回", fmt.Sprintf("%d-back", ID)),
		tgbotapi.NewInlineKeyboardButtonData("🔚 退出", fmt.Sprintf("%d-exit", ID)),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, row4, row5, row6)
	return &keyboard
}
