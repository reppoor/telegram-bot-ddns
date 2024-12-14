package keyboard

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func createInlineKeyboard(keyboard InlineKeyboard) tgbotapi.InlineKeyboardMarkup {
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

	// 创建退出按钮并添加到最后一行
	exitButton := tgbotapi.NewInlineKeyboardButtonData("退出🔚", "1-exit")
	rows = append(rows, []tgbotapi.InlineKeyboardButton{exitButton})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// GenerateMainMenuKeyboard 生成一级菜单按钮
func GenerateMainMenuKeyboard(domainMap map[string]map[string]map[string]interface{}) tgbotapi.InlineKeyboardMarkup {
	var keyboard InlineKeyboard

	for domainName, forwardingMap := range domainMap {
		for forwardingDomain, details := range forwardingMap {
			// 提取端口信息并格式化按钮文本
			port := details["Port"]
			ban, _ := details["Ban"].(bool)
			buttonText := fmt.Sprintf("%s - %s - %v - %t", domainName, forwardingDomain, port, ban)

			// 将回调数据设置为例如 ID
			callbackData := fmt.Sprintf("%v", details["ID"])

			// 创建按钮
			button := Button{
				Text:         buttonText,
				CallbackData: callbackData,
			}

			// 将每个按钮作为单独一行（竖向排列）
			keyboard.Buttons = append(keyboard.Buttons, []Button{button})
		}
	}

	return createInlineKeyboard(keyboard)
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
