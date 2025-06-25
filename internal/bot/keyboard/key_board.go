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
		// 设置按钮文本
		BanText := "✅️"
		if domain.Ban {
			BanText = "❌️️"
		}
		buttonText := fmt.Sprintf("%s - %s - %d - %s", domain.Domain, domain.ForwardingDomain, domain.Port, BanText)
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
		delText := "🚫️"
		if domain.Del {
			delText = "✅️️"
		}

		buttonText := fmt.Sprintf("%s - %s - %s - %d", delText, domain.Domain, domain.ForwardingDomain, domain.Port)
		callbackData := fmt.Sprintf("%d-delete", domain.ID)

		button := Button{
			Text:         buttonText,
			CallbackData: callbackData,
		}

		// 每个按钮单独一行
		keyboard.Buttons = append(keyboard.Buttons, []Button{button})
	}

	return createInlineKeyboard(keyboard, 2)
}

// GenerateSubMenuKeyboard 生成二级菜单按钮
func GenerateSubMenuKeyboard(ID uint, Ban bool) *tgbotapi.InlineKeyboardMarkup {
	// 设置按钮文本
	BanText := "启用中✅️"
	if Ban {
		BanText = "已封禁🚫️️"
	}

	// 定义按钮
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(BanText, fmt.Sprintf("%d-ban", ID)),
		tgbotapi.NewInlineKeyboardButtonData("设置权重", fmt.Sprintf("%d-weight", ID)),
		tgbotapi.NewInlineKeyboardButtonData("获取转发最新IP🔝", fmt.Sprintf("%d-getIp", ID)),
		tgbotapi.NewInlineKeyboardButtonData("解析该条记录📶", fmt.Sprintf("%d-parse", ID)),
		tgbotapi.NewInlineKeyboardButtonData("检测并解析该条记录🔄", fmt.Sprintf("%d-checkAndParse", ID)),
		tgbotapi.NewInlineKeyboardButtonData("删除该条记录❌️", fmt.Sprintf("%d-del", ID)),
		tgbotapi.NewInlineKeyboardButtonData("返回🔙", fmt.Sprintf("%d-back", ID)),
		tgbotapi.NewInlineKeyboardButtonData("退出🔚", fmt.Sprintf("%d-exit", ID)),
	}

	// 创建竖直排列的键盘
	var inlineRows [][]tgbotapi.InlineKeyboardButton
	for _, button := range buttons {
		inlineRows = append(inlineRows, []tgbotapi.InlineKeyboardButton{button})
	}

	// 返回键盘布局
	keyboard := tgbotapi.NewInlineKeyboardMarkup(inlineRows...)
	return &keyboard
}
