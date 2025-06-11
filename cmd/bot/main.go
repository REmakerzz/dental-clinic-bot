package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/REmakerzz/dental-clinic-bot/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New()
	if err != nil {
		log.Fatalf("❌ Failed to start app: %v", err)
	}
	defer a.Close()
	a.Run(ctx)
}