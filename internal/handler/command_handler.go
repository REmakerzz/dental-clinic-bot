package handler

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/REmakerzz/dental-clinic-bot/internal/config"
	"github.com/REmakerzz/dental-clinic-bot/internal/model"
	"github.com/REmakerzz/dental-clinic-bot/internal/service"
	"github.com/REmakerzz/dental-clinic-bot/internal/ui"
)

type CommandHandler struct {
	bot            *tgbotapi.BotAPI
	groupChatID    int64
	userBookings   map[int64]*model.Booking
	bookingService *service.BookingService
	config         *config.Config
}

func NewCommandHandler(bot *tgbotapi.BotAPI, groupChatID int64, bookingService *service.BookingService, userBookings map[int64]*model.Booking, cfg *config.Config) *CommandHandler {
	return &CommandHandler{
		bot:            bot,
		groupChatID:    groupChatID,
		userBookings:   userBookings,
		bookingService: bookingService,
		config:         cfg,
	}
}

func (h *CommandHandler) HandleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	if msg.IsCommand() {
		switch msg.Command() {
		case "admin_help":
			h.handleAdminHelp(chatID, msg.From.ID)

		case "admin_list":
			h.handleAdminList(chatID, msg.From.ID)

		case "admin_stats":
			h.handleAdminStats(chatID, msg.From.ID)

		case "admin_delete":
			h.handleAdminDelete(chatID, msg.From.ID, msg.CommandArguments())

		default:
			h.bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда."))
		}
	} else {
		// некомандные сообщения → можно потом сюда добавить обработку
		h.handleBookingFlow(msg)
	}
}

func (h *CommandHandler) handleAdminHelp(chatID int64, userID int64) {
	if service.IsAdmin(userID, h.config.AdminUserIDs) {
		helpText := "Доступные админ-команды:\n\n" +
			"/admin_list — Показать все заявки\n" +
			"/admin_stats — Показать статистику\n" +
			"/admin_delete N — Удалить заявку по ID\n" +
			"/admin_help — Показать это сообщение\n"

		h.bot.Send(tgbotapi.NewMessage(chatID, helpText))
	} else {
		h.bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для этой команды."))
	}
}

func (h *CommandHandler) handleAdminList(chatID int64, userID int64) {
	if service.IsAdmin(userID, h.config.AdminUserIDs) {
		bookings, err := h.bookingService.GetAllBookings()
		if err != nil {
			h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка получения заявок."))
			return
		}

		if len(bookings) == 0 {
			h.bot.Send(tgbotapi.NewMessage(chatID, "Заявок пока нет."))
			return
		}

		for _, b := range bookings {
			text := "ID: " + strconv.Itoa(b.ID) + "\n" +
				"Имя: " + b.Name + "\n" +
				"Телефон: " + b.Phone + "\n" +
				"Услуга: " + b.Service + "\n" +
				"Дата и время: " + b.DateTime

			deleteButton := tgbotapi.NewInlineKeyboardButtonData("❌ Удалить заявку", "delete:"+strconv.Itoa(b.ID))
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(deleteButton),
			)

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ReplyMarkup = keyboard

			h.bot.Send(msg)
		}
	} else {
		h.bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для этой команды."))
	}
}

func (h *CommandHandler) handleAdminStats(chatID int64, userID int64) {
	if service.IsAdmin(userID, h.config.AdminUserIDs) {
		total, today, last7Days, err := h.bookingService.GetBookingStats()
		if err != nil {
			h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка получения статистики."))
			return
		}

		statsText := "📊 Статистика заявок:\n\n" +
			"Всего заявок: " + strconv.Itoa(total) + "\n" +
			"Заявок сегодня: " + strconv.Itoa(today) + "\n" +
			"Заявок за последние 7 дней: " + strconv.Itoa(last7Days)

		h.bot.Send(tgbotapi.NewMessage(chatID, statsText))
	} else {
		h.bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для этой команды."))
	}
}

func (h *CommandHandler) handleAdminDelete(chatID int64, userID int64, args string) {
	if service.IsAdmin(userID, h.config.AdminUserIDs) {
		id, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			h.bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, укажите корректный ID заявки: /admin_delete 123"))
			return
		}

		err = h.bookingService.DeleteBookingByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				h.bot.Send(tgbotapi.NewMessage(chatID, "Заявка с таким ID не найдена."))
			} else {
				h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка удаления заявки."))
			}
			return
		}

		h.bot.Send(tgbotapi.NewMessage(chatID, "Заявка успешно удалена."))
	} else {
		h.bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для этой команды."))
	}
}

func (h *CommandHandler) handleBookingFlow(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	// стартуем процесс
	if text == "🗓️ Записаться на приём" {
		h.userBookings[chatID] = &model.Booking{Step: 1}
		msg := tgbotapi.NewMessage(chatID, "Как вас зовут?")
		h.bot.Send(msg)
		return
	}

	booking, exists := h.userBookings[chatID]
	if exists {
		switch booking.Step {
		case 1:
			booking.Name = text
			booking.Step++
			msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введите ваш номер телефона:")
			h.bot.Send(msg)
		case 2:
			booking.Phone = text
			booking.Step++
			msg := tgbotapi.NewMessage(chatID, "Какую услугу вы хотите получить?")
			msg.ReplyMarkup = ui.ServiceKeyboard()
			h.bot.Send(msg)
		case 3:
			booking.Service = text
			booking.Step++
			msg := tgbotapi.NewMessage(chatID, "На какую дату вы хотите записаться? (формат: YYYY-MM-DD)")
			h.bot.Send(msg)
		case 4:
			// Validate date format
			_, err := time.Parse("2006-01-02", text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Неверный формат даты. Пожалуйста, используйте формат YYYY-MM-DD")
				h.bot.Send(msg)
				return
			}

			// Get available time slots for the selected date
			slots, err := h.bookingService.GetAvailableTimeSlots(text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Ошибка при получении доступного времени. Пожалуйста, попробуйте другую дату.")
				h.bot.Send(msg)
				return
			}

			if len(slots) == 0 {
				msg := tgbotapi.NewMessage(chatID, "На выбранную дату нет доступного времени. Пожалуйста, выберите другую дату.")
				h.bot.Send(msg)
				return
			}

			// Create keyboard with available time slots
			var keyboard [][]tgbotapi.InlineKeyboardButton
			for _, slot := range slots {
				timeStr := slot[11:16] // Extract time part (HH:MM)
				keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData(timeStr, "time:"+slot),
				})
			}

			msg := tgbotapi.NewMessage(chatID, "Выберите удобное время:")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
			h.bot.Send(msg)
			booking.Step++
		case 5:
			// This step is handled by callback handler
			return
		default:
			msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие из меню.")
			msg.ReplyMarkup = ui.MainMenuKeyboard()
			h.bot.Send(msg)
		}

		return
	}

	// если просто сообщение → показываем соответствующее меню
	if service.IsAdmin(msg.From.ID, h.config.AdminUserIDs) || chatID == h.groupChatID {
		msg := tgbotapi.NewMessage(chatID, "Администратор пожалуйста выберите действие:")
		msg.ReplyMarkup = ui.AdminMenuKeyboard()
		h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, "Привет! Это Денталь бот стоматологической клиники. Выберите действие:")
		msg.ReplyMarkup = ui.MainMenuKeyboard()
		h.bot.Send(msg)
	}
}
