package handler

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/REmakerzz/dental-clinic-bot/internal/config"
	"github.com/REmakerzz/dental-clinic-bot/internal/model"
	"github.com/REmakerzz/dental-clinic-bot/internal/service"
	"github.com/REmakerzz/dental-clinic-bot/internal/ui"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler struct {
	bot            *tgbotapi.BotAPI
	bookingService *service.BookingService
	config         *config.Config
	userBookings   map[int64]*model.Booking
}

func NewCallbackHandler(bot *tgbotapi.BotAPI, bookingService *service.BookingService, config *config.Config, userBookings map[int64]*model.Booking) *CallbackHandler {
	return &CallbackHandler{
		bot:            bot,
		bookingService: bookingService,
		config:         config,
		userBookings:   userBookings,
	}
}

func (h *CallbackHandler) HandleCallback(callback *tgbotapi.CallbackQuery) {
	data := callback.Data

	if strings.HasPrefix(data, "delete:") {
		h.handleDeleteCallback(callback, data)
	} else if strings.HasPrefix(data, "time:") {
		h.handleTimeSelection(callback, data)
	} else {
		callbackResp := tgbotapi.NewCallback(callback.ID, "Неизвестный callback.")
		h.bot.Request(callbackResp)
	}
}

func (h *CallbackHandler) handleTimeSelection(callback *tgbotapi.CallbackQuery, data string) {
	chatID := callback.Message.Chat.ID
	datetime := strings.TrimPrefix(data, "time:")

	// Get the booking from the map
	booking, exists := h.userBookings[chatID]
	if !exists {
		callbackResp := tgbotapi.NewCallback(callback.ID, "Ошибка: сессия бронирования не найдена.")
		h.bot.Request(callbackResp)
		return
	}

	// Set the datetime
	booking.DateTime = datetime

	// Save the booking
	err := h.bookingService.SaveBooking(booking)
	if err != nil {
		callbackResp := tgbotapi.NewCallback(callback.ID, "Ошибка при сохранении записи.")
		h.bot.Request(callbackResp)
		return
	}

	// Send confirmation to user
	confirmMsg := tgbotapi.NewMessage(chatID, "Спасибо за запись! Заявка сохранена.")
	confirmMsg.ReplyMarkup = ui.MainMenuKeyboard()
	h.bot.Send(confirmMsg)

	// Send notification to admin
	adminMsg := tgbotapi.NewMessage(h.config.AdminGroupChatID, "Новая запись на приём:\n\n"+
		"Имя: "+booking.Name+"\n"+
		"Телефон: "+booking.Phone+"\n"+
		"Услуга: "+booking.Service+"\n"+
		"Дата и время: "+booking.DateTime)
	h.bot.Send(adminMsg)

	// Delete the booking from the map
	delete(h.userBookings, chatID)

	// Delete the time selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	h.bot.Request(deleteMsg)
}

func (h *CallbackHandler) handleDeleteCallback(callback *tgbotapi.CallbackQuery, data string) {
	if !service.IsAdmin(callback.From.ID, h.config.AdminUserIDs) {
		callbackResp := tgbotapi.NewCallback(callback.ID, "У вас нет прав для этой операции.")
		h.bot.Request(callbackResp)
		return
	}

	idStr := strings.TrimPrefix(data, "delete:")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(callback.Message.Chat.ID, "Некорректный ID заявки."))
		return
	}

	err = h.bookingService.DeleteBookingByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			h.bot.Send(tgbotapi.NewMessage(callback.Message.Chat.ID, "Заявка с таким ID не найдена."))
		} else {
			h.bot.Send(tgbotapi.NewMessage(callback.Message.Chat.ID, "Ошибка удаления заявки."))
		}
	} else {
		// Уведомление Telegram, что все ок
		callbackResp := tgbotapi.NewCallback(callback.ID, "Заявка успешно удалена.")
		h.bot.Request(callbackResp)

		// Удаление сообщения с заявкой
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		h.bot.Request(deleteMsg)
	}
}
