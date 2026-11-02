package services

import (
	"context"
	"log"
	"time"

	"qabulgo-backend/internal/repositories"
)

type NotificationWorker interface {
	Start(ctx context.Context)
}

type notificationWorker struct {
	outboxRepo      repositories.OutboxRepository
	telegramService TelegramService
	pollInterval    time.Duration
}

func NewNotificationWorker(
	outboxRepo repositories.OutboxRepository,
	telegramService TelegramService,
) NotificationWorker {
	return &notificationWorker{
		outboxRepo:      outboxRepo,
		telegramService: telegramService,
		pollInterval:    3 * time.Second,
	}
}

func (w *notificationWorker) Start(ctx context.Context) {
	go func() {
		log.Println("Notification Outbox Worker ishga tushdi...")
		ticker := time.NewTicker(w.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Notification Outbox Worker to‘xtatildi.")
				return
			case <-ticker.C:
				w.processPending(ctx)
			}
		}
	}()
}

func (w *notificationWorker) processPending(ctx context.Context) {
	items, err := w.outboxRepo.GetPending(ctx, 20)
	if err != nil {
		return
	}

	for _, item := range items {
		select {
		case <-ctx.Done():
			return
		default:
			if err := w.telegramService.DispatchOutboxNotification(ctx, &item); err != nil {
				log.Printf("Outbox notification yuborishda xatolik (ID: %s): %v\n", item.ID, err)
			}
		}
	}
}

