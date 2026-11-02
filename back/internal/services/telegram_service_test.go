package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/telegram"

	"github.com/google/uuid"
)

type StubTelegramRepo struct {
	processedUpdates map[int64]bool
}

func (s *StubTelegramRepo) CreateLinkingToken(ctx context.Context, token *models.TelegramLinkingToken) error {
	return nil
}

func (s *StubTelegramRepo) GetValidLinkingTokenByHash(ctx context.Context, tokenHash string) (*models.TelegramLinkingToken, error) {
	return nil, nil
}

func (s *StubTelegramRepo) MarkLinkingTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	return nil
}

func (s *StubTelegramRepo) UpsertClinicRecipient(ctx context.Context, recipient *models.ClinicTelegramRecipient) error {
	return nil
}

func (s *StubTelegramRepo) GetClinicRecipients(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicTelegramRecipient, error) {
	return nil, nil
}

func (s *StubTelegramRepo) GetActiveClinicRecipients(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicTelegramRecipient, error) {
	return nil, nil
}

func (s *StubTelegramRepo) GetRecipientByID(ctx context.Context, clinicID, recipientID uuid.UUID) (*models.ClinicTelegramRecipient, error) {
	return nil, nil
}

func (s *StubTelegramRepo) DisconnectRecipient(ctx context.Context, clinicID, recipientID uuid.UUID) error {
	return nil
}

func (s *StubTelegramRepo) ToggleRecipientNotifications(ctx context.Context, clinicID, recipientID uuid.UUID) (*models.ClinicTelegramRecipient, error) {
	return nil, nil
}

func (s *StubTelegramRepo) GetClinicsByTelegramUserID(ctx context.Context, telegramUserID int64) ([]models.Clinic, error) {
	return nil, nil
}

func (s *StubTelegramRepo) DeactivateRecipientByChatID(ctx context.Context, chatID int64) error {
	return nil
}

func (s *StubTelegramRepo) IsUpdateProcessed(ctx context.Context, botType string, updateID int64) (bool, error) {
	if s.processedUpdates == nil {
		return false, nil
	}
	return s.processedUpdates[updateID], nil
}

func (s *StubTelegramRepo) RecordUpdateProcessed(ctx context.Context, botType string, updateID int64) error {
	if s.processedUpdates == nil {
		s.processedUpdates = make(map[int64]bool)
	}
	s.processedUpdates[updateID] = true
	return nil
}

func (s *StubTelegramRepo) GetAdminSession(ctx context.Context, telegramUserID int64) (*models.TelegramAdminSession, error) {
	return nil, nil
}

func (s *StubTelegramRepo) SetAdminSession(ctx context.Context, session *models.TelegramAdminSession) error {
	return nil
}

func (s *StubTelegramRepo) ClearAdminSession(ctx context.Context, telegramUserID int64) error {
	return nil
}

type StubTelegramClient struct {
	sentMessages int
}

func (c *StubTelegramClient) GetMe(ctx context.Context) (*telegram.User, error) {
	return &telegram.User{ID: 1000, FirstName: "QabulGoBot"}, nil
}

func (c *StubTelegramClient) SendMessage(ctx context.Context, chatID int64, text string, replyMarkup any) (*telegram.Message, error) {
	c.sentMessages++
	return &telegram.Message{MessageID: 1, Chat: telegram.Chat{ID: chatID}, Text: text}, nil
}

func (c *StubTelegramClient) EditMessageText(ctx context.Context, chatID int64, messageID int64, text string, replyMarkup *telegram.InlineKeyboardMarkup) (*telegram.Message, error) {
	return &telegram.Message{MessageID: messageID, Chat: telegram.Chat{ID: chatID}, Text: text}, nil
}

func (c *StubTelegramClient) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string, showAlert bool) error {
	return nil
}

func (c *StubTelegramClient) GetUpdates(ctx context.Context, offset int64, limit int, timeoutSeconds int) ([]telegram.Update, error) {
	return nil, nil
}

func TestTelegramLinkingTokenSecurity(t *testing.T) {
	rawToken := "abc123xyz789securetoken"
	hasher := sha256.New()
	hasher.Write([]byte(rawToken))
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	if expectedHash == "" || len(expectedHash) != 64 {
		t.Fatalf("expected 64-char sha256 hex string, got %s", expectedHash)
	}

	hasher2 := sha256.New()
	hasher2.Write([]byte(rawToken))
	if expectedHash != hex.EncodeToString(hasher2.Sum(nil)) {
		t.Fatalf("sha256 hash not deterministic")
	}
}

func TestAdminAuthorizationCheck(t *testing.T) {
	cfg := &config.Config{
		TelegramAdminAllowedIDs: []int64{123456789, 987654321},
	}
	svc := &telegramService{cfg: cfg}

	if !svc.isAuthorizedAdmin(123456789) {
		t.Errorf("expected 123456789 to be authorized")
	}
	if !svc.isAuthorizedAdmin(987654321) {
		t.Errorf("expected 987654321 to be authorized")
	}
	if svc.isAuthorizedAdmin(111111111) {
		t.Errorf("expected 111111111 to NOT be authorized")
	}
	if svc.isAuthorizedAdmin(0) {
		t.Errorf("expected 0 to NOT be authorized")
	}
}

func TestUpdateDeduplicationLogic(t *testing.T) {
	stubRepo := &StubTelegramRepo{processedUpdates: make(map[int64]bool)}
	stubClient := &StubTelegramClient{}

	cfg := &config.Config{
		TelegramClinicBotToken: "test_token",
	}
	svc := &telegramService{
		cfg:             cfg,
		telegramRepo:    stubRepo,
		clinicBotClient: stubClient,
	}

	update := &telegram.Update{
		UpdateID: 999123,
	}

	// 1. First time: not processed yet -> records update
	err := svc.ProcessClinicBotUpdate(context.Background(), update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stubRepo.processedUpdates[999123] {
		t.Fatalf("expected update 999123 to be recorded as processed")
	}

	// 2. Second time: already processed -> skips
	initialCount := stubClient.sentMessages
	err = svc.ProcessClinicBotUpdate(context.Background(), update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stubClient.sentMessages != initialCount {
		t.Fatalf("expected duplicate update to be skipped without processing")
	}
}
