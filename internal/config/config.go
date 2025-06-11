package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken    string
	AdminGroupChatID int64
	AdminUserIDs     []int64
}

func LoadConfig() (*Config, error) {
	// Загружаем .env (если есть)
	_ = godotenv.Load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
	}

	groupChatStr := os.Getenv("ADMIN_GROUP_CHAT_ID")
	groupChatID, err := strconv.ParseInt(groupChatStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ADMIN_GROUP_CHAT_ID: %w", err)
	}

	adminsRaw := os.Getenv("ADMIN_USER_IDS")
	if adminsRaw == "" {
		return nil, fmt.Errorf("ADMIN_USER_IDS is not set")
	}

	adminIDs := []int64{}
	for _, idStr := range strings.Split(adminsRaw, ",") {
		idStr = strings.TrimSpace(idStr)
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid admin ID: %s", idStr)
		}
		adminIDs = append(adminIDs, id)
	}

	return &Config{
		TelegramToken:    token,
		AdminGroupChatID: groupChatID,
		AdminUserIDs:     adminIDs,
	}, nil
}