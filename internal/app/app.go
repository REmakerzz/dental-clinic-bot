package app

import (
	"context"
	"database/sql"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/REmakerzz/dental-clinic-bot/internal/config"
	"github.com/REmakerzz/dental-clinic-bot/internal/handler"
	"github.com/REmakerzz/dental-clinic-bot/internal/model"
	"github.com/REmakerzz/dental-clinic-bot/internal/repository"
	"github.com/REmakerzz/dental-clinic-bot/internal/service"
)

type App struct {
	bot             *tgbotapi.BotAPI
	commandHandler  *handler.CommandHandler
	callbackHandler *handler.CallbackHandler
	config          *config.Config
	db              *sql.DB
	userBookings    map[int64]*model.Booking
}

func New() (*App, error) {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Init DB
	db := repository.InitDB()
	// Не закрываем здесь — defer в main.go (или можно использовать App.Close позже)

	// Init services
	bookingService := service.NewBookingService(db)

	// Init bot
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Bot authorized as @%s", bot.Self.UserName)
	log.Printf("📦 AdminGroup: %d | Admins: %v", cfg.AdminGroupChatID, cfg.AdminUserIDs)

	// Create shared userBookings map
	userBookings := make(map[int64]*model.Booking)

	// Init handlers
	commandHandler := handler.NewCommandHandler(bot, cfg.AdminGroupChatID, bookingService, userBookings, cfg)
	callbackHandler := handler.NewCallbackHandler(bot, bookingService, cfg, userBookings)

	return &App{
		bot:             bot,
		commandHandler:  commandHandler,
		callbackHandler: callbackHandler,
		config:          cfg,
		db:              db,
		userBookings:    userBookings,
	}, nil
}

func (a *App) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := a.bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			if update.Message != nil {
				a.commandHandler.HandleMessage(update.Message)
			} else if update.CallbackQuery != nil {
				a.callbackHandler.HandleCallback(update.CallbackQuery)
			}
		case <-ctx.Done():
			log.Println("🔌 Shutdown signal received. Stopping bot...")
			time.Sleep(1 * time.Second)
			return
		}
	}
}

func (a *App) Close() {
	log.Println("Closing database...")
	if err := a.db.Close(); err != nil {
		log.Printf("Error closing DB: %v", err)
	}
}
