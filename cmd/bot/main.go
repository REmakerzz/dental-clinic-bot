package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/REmakerzz/dental-clinic-bot/internal/model"
	"github.com/REmakerzz/dental-clinic-bot/internal/repository"
	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	adminUserIDs []int64
	db           *sql.DB
	userBookings = make(map[int64]*model.Booking)
)

func main() {

	db = repository.InitDB()
	defer db.Close()

	//Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	groupChatIDStr := os.Getenv("ADMIN_GROUP_CHAT_ID")
	groupChatID, err := strconv.ParseInt(groupChatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid ADMIN_GROUP_CHAT_ID: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	loadAdminUserIDs()
	log.Printf("Loaded adminUserIDs: %v", adminUserIDs)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID

			// если новое сообщение "Записаться на приём" — стартуем процесс
			if update.Message.Text == "🗓️ Записаться на приём" {
				userBookings[chatID] = &model.Booking{Step: 1}
				msg := tgbotapi.NewMessage(chatID, "Как вас зовут?")
				bot.Send(msg)
				continue
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "admin_help":
					if isAdmin(update.Message.From.ID) {
						helpText := "Доступные админ-команды:\n\n" +
							"/admin_list — Показать все заявки\n" +
							"/admin_stats — Показать статистику\n" +
							"/admin_help — Показать это сообщение\n"

						msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
						msg.ReplyMarkup = adminMenuKeyboard()
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для этой команды.")
						bot.Send(msg)
					}
					continue
				case "admin_list":
					if isAdmin(update.Message.From.ID) {
						bookings, err := repository.GetAllBookings(db)
						if err != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения заявок.")
							bot.Send(msg)
							continue
						}

						if len(bookings) == 0 {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заявок пока нет.")
							bot.Send(msg)
							continue
						}

						const batchSize = 10
						for i := 0; i < len(bookings); i += batchSize {
							end := i + batchSize
							if end > len(bookings) {
								end = len(bookings)
							}

							text := "Список заявок:\n\n"
							for _, b := range bookings[i:end] {
								text += "ID: " + strconv.Itoa(b.ID) + "\n" +
									"Имя: " + b.Name + "\n" +
									"Телефон: " + b.Phone + "\n" +
									"Услуга: " + b.Service + "\n" +
									"Дата и время: " + b.DateTime + "\n---\n"
							}

							msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
							bot.Send(msg)
						}
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для этой команды.")
						bot.Send(msg)
					}
					continue
				case "admin_stats":
					if isAdmin(update.Message.From.ID) {
						total, today, last7Days, err := repository.GetBookingStats(db)
						if err != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения статистики.")
							bot.Send(msg)
							continue
						}

						statsText := "📊 Статистика заявок:\n\n" +
							"Всего заявок: " + strconv.Itoa(total) + "\n" +
							"Заявок сегодня: " + strconv.Itoa(today) + "\n" +
							"Заявок за последние 7 дней: " + strconv.Itoa(last7Days)

						msg := tgbotapi.NewMessage(update.Message.Chat.ID, statsText)
						msg.ReplyMarkup = adminMenuKeyboard()
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для этой команды.")
						bot.Send(msg)
					}
					continue
				case "admin_delete":
					if isAdmin(update.Message.From.ID) {
						args := update.Message.CommandArguments()
						id, err := strconv.Atoi(strings.TrimSpace(args))
						if err != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите корректный ID заявки: /admin_delete 123")
							bot.Send(msg)
							continue
						}

						err = repository.DeleteBookingByID(db, id)
						if err != nil {
							if err == sql.ErrNoRows {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заявка с таким ID не найдена.")
								bot.Send(msg)
							} else {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка удаления заявки.")
								bot.Send(msg)
							}
							continue
						}

						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заявка успешно удалена.")
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для этой команды.")
						bot.Send(msg)
					}
					continue
				}
			}

			booking, exists := userBookings[chatID]
			if exists {
				switch booking.Step {
				case 1:
					booking.Name = update.Message.Text
					booking.Step++
					msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введите ваш номер телефона:")
					bot.Send(msg)
				case 2:
					booking.Phone = update.Message.Text
					booking.Step++
					msg := tgbotapi.NewMessage(chatID, "Какую услугу вы хотите получить?")
					msg.ReplyMarkup = serviceKeyboard()
					bot.Send(msg)
				case 3:
					booking.Service = update.Message.Text
					booking.Step++
					msg := tgbotapi.NewMessage(chatID, "На какую дату и время вы хотите записаться?")
					bot.Send(msg)
				case 4:
					booking.DateTime = update.Message.Text
					booking.Step = 0 // сбросим step

					// Сохраняем в БД
					err := repository.SaveBooking(db, booking)
					if err != nil {
						log.Printf("Error saving booking: %v", err)
					}

					// Отправляем подтверждение пользователю
					confirmMsg := tgbotapi.NewMessage(chatID, "Спасибо за запись! Заявка сохранена.")
					bot.Send(confirmMsg)

					// Отправляем админу
					adminMsg := tgbotapi.NewMessage(groupChatID, "Новая запись на приём:\n\n"+
						"Имя: "+booking.Name+"\n"+
						"Телефон: "+booking.Phone+"\n"+
						"Услуга: "+booking.Service+"\n"+
						"Дата и время: "+booking.DateTime)
					bot.Send(adminMsg)

					// Удаляем из map
					delete(userBookings, chatID)
				default:
					msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие из меню.")
					msg.ReplyMarkup = mainMenuKeyboard()
					bot.Send(msg)
				}

				continue
			}

			// если просто сообщение → главное меню
			// если просто сообщение → показываем соответствующее меню
			if isAdmin(update.Message.From.ID) || chatID == groupChatID {
				msg := tgbotapi.NewMessage(chatID, "Администратор пожалуйста выберите действие:")
				msg.ReplyMarkup = adminMenuKeyboard()
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Привет! Это бот стоматологической клиники. Выберите действие:")
				msg.ReplyMarkup = mainMenuKeyboard()
				bot.Send(msg)
			}
		}
	}
}

func mainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
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

func adminMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
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

func serviceKeyboard() tgbotapi.ReplyKeyboardMarkup {
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

func loadAdminUserIDs() {
	idsStr := os.Getenv("ADMIN_USER_IDS")
	idsSlice := strings.Split(idsStr, ",")
	for _, idStr := range idsSlice {
		idStr = strings.TrimSpace(idStr)
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Fatalf("Invalid ADMIN__USER_IDS: %v", err)
		}
		adminUserIDs = append(adminUserIDs, idInt)
	}
}

func isAdmin(userID int64) bool {
	for _, id := range adminUserIDs {
		if userID == id {
			return true
		}
	}
	return false
}
