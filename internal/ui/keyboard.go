package ui

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func MainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
    keyboard := tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("🗓️ Записаться на приём"),
            tgbotapi.NewKeyboardButton("📋 Наши услуги"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("💳 Цены"),
            tgbotapi.NewKeyboardButton("📞 Контакты"),
        ),
    )
    keyboard.ResizeKeyboard = true
    return keyboard
}

func AdminMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
    keyboard := tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("/admin_list"),
            tgbotapi.NewKeyboardButton("/admin_stats"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("/admin_help"),
            tgbotapi.NewKeyboardButton("↩️ Главное меню"),
        ),
    )
    keyboard.ResizeKeyboard = true
    return keyboard
}

func ServiceKeyboard() tgbotapi.ReplyKeyboardMarkup {
    keyboard := tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Профессиональная чистка"),
            tgbotapi.NewKeyboardButton("Лечение кариеса"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Протезирование"),
            tgbotapi.NewKeyboardButton("Имплантация"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Отбеливание"),
            tgbotapi.NewKeyboardButton("Другое"),
        ),
    )
    keyboard.ResizeKeyboard = true
    return keyboard
}