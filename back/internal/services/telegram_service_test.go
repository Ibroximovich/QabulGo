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
	adminTokens      map[string]*models.AdminTelegramLinkingToken
	adminRecipients  map[int64]*models.AdminTelegramRecipient
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

func (s *StubTelegramRepo) CreateAdminLinkingToken(ctx context.Context, token *models.AdminTelegramLinkingToken) error {
	if s.adminTokens == nil {
		s.adminTokens = make(map[string]*models.AdminTelegramLinkingToken)
	}
	s.adminTokens[token.TokenHash] = token
	return nil
}

func (s *StubTelegramRepo) GetValidAdminLinkingTokenByHash(ctx context.Context, tokenHash string) (*models.AdminTelegramLinkingToken, error) {
	if s.adminTokens == nil {
		return nil, nil
	}
	token := s.adminTokens[tokenHash]
	if token != nil && !token.IsUsed {
		return token, nil
	}
	return nil, nil
}

func (s *StubTelegramRepo) MarkAdminLinkingTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	for _, tok := range s.adminTokens {
		if tok.ID == tokenID {
			tok.IsUsed = true
		}
	}
	return nil
}

func (s *StubTelegramRepo) UpsertAdminRecipient(ctx context.Context, recipient *models.AdminTelegramRecipient) error {
	if s.adminRecipients == nil {
		s.adminRecipients = make(map[int64]*models.AdminTelegramRecipient)
	}
	s.adminRecipients[recipient.TelegramUserID] = recipient
	return nil
}

func (s *StubTelegramRepo) GetAdminRecipientByUserID(ctx context.Context, userID uuid.UUID) (*models.AdminTelegramRecipient, error) {
	for _, r := range s.adminRecipients {
		if r.UserID == userID && r.IsActive {
			return r, nil
		}
	}
	return nil, nil
}

func (s *StubTelegramRepo) GetAdminRecipientByTelegramUserID(ctx context.Context, telegramUserID int64) (*models.AdminTelegramRecipient, error) {
	if s.adminRecipients == nil {
		return nil, nil
	}
	return s.adminRecipients[telegramUserID], nil
}

func (s *StubTelegramRepo) GetAllActiveAdminRecipients(ctx context.Context) ([]models.AdminTelegramRecipient, error) {
	var list []models.AdminTelegramRecipient
	for _, r := range s.adminRecipients {
		if r.IsActive {
			list = append(list, *r)
		}
	}
	return list, nil
}

func (s *StubTelegramRepo) DisconnectAdminRecipient(ctx context.Context, userID uuid.UUID) error {
	for _, r := range s.adminRecipients {
		if r.UserID == userID {
			r.IsActive = false
		}
	}
	return nil
}

func (s *StubTelegramRepo) DeactivateAdminRecipientByChatID(ctx context.Context, chatID int64) error {
	for _, r := range s.adminRecipients {
		if r.TelegramChatID == chatID {
			r.IsActive = false
		}
	}
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
	ctx := context.Background()
	cfg := &config.Config{
		TelegramAdminAllowedIDs: []int64{123456789, 987654321},
	}
	stubRepo := &StubTelegramRepo{}
	svc := &telegramService{cfg: cfg, telegramRepo: stubRepo}

	if !svc.isAuthorizedAdmin(ctx, 123456789) {
		t.Errorf("expected 123456789 to be authorized")
	}
	if !svc.isAuthorizedAdmin(ctx, 987654321) {
		t.Errorf("expected 987654321 to be authorized")
	}
	if svc.isAuthorizedAdmin(ctx, 111111111) {
		t.Errorf("expected 111111111 to NOT be authorized")
	}
	if svc.isAuthorizedAdmin(ctx, 0) {
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

func TestAdminLinkingTokenFlow(t *testing.T) {
	ctx := context.Background()
	stubRepo := &StubTelegramRepo{}
	stubClient := &StubTelegramClient{}

	cfg := &config.Config{
		TelegramAdminBotToken:    "admin_test_token",
		TelegramAdminBotUsername: "qabulgo_admin_bot",
	}

	svc := &telegramService{
		cfg:             cfg,
		telegramRepo:    stubRepo,
		adminBotClient:  stubClient,
	}

	adminUserID := uuid.New()

	// 1. Generate Linking Token
	resp, err := svc.GenerateAdminLinkingToken(ctx, adminUserID)
	if err != nil {
		t.Fatalf("GenerateAdminLinkingToken failed: %v", err)
	}

	if resp.Token == "" || resp.DeepLink == "" {
		t.Fatalf("expected non-empty token and deep link")
	}

	// 2. Before linking: unlinked telegram ID 1100274830 is NOT authorized
	telegramUserID := int64(1100274830)
	if svc.isAuthorizedAdmin(ctx, telegramUserID) {
		t.Fatalf("expected telegram ID %d to NOT be authorized before linking", telegramUserID)
	}

	// 3. Process /start link_<token> update
	update := &telegram.Update{
		UpdateID: 1001,
		Message: &telegram.Message{
			MessageID: 10,
			Chat:      telegram.Chat{ID: 1100274830, Type: "private"},
			From: &telegram.User{
				ID:        telegramUserID,
				Username:  "ibroximjon_uz",
				FirstName: "Ibrohimjon",
			},
			Text: "/start link_" + resp.Token,
		},
	}

	err = svc.ProcessAdminBotUpdate(ctx, update)
	if err != nil {
		t.Fatalf("ProcessAdminBotUpdate failed: %v", err)
	}

	// 4. Verify recipient was created in repo and is now authorized
	if !svc.isAuthorizedAdmin(ctx, telegramUserID) {
		t.Fatalf("expected telegram ID %d to be authorized after successful linking", telegramUserID)
	}

	// 5. Status check
	status, err := svc.GetAdminTelegramStatus(ctx, adminUserID)
	if err != nil {
		t.Fatalf("GetAdminTelegramStatus failed: %v", err)
	}
	if !status.IsConnected {
		t.Fatalf("expected IsConnected to be true")
	}
	if status.ConnectedAdmin == nil || status.ConnectedAdmin.TelegramUserID != telegramUserID {
		t.Fatalf("expected ConnectedAdmin.TelegramUserID to be %d", telegramUserID)
	}

	// 6. Disconnect
	err = svc.DisconnectAdminRecipient(ctx, adminUserID)
	if err != nil {
		t.Fatalf("DisconnectAdminRecipient failed: %v", err)
	}

	if svc.isAuthorizedAdmin(ctx, telegramUserID) {
		t.Fatalf("expected telegram ID %d to NOT be authorized after disconnect", telegramUserID)
	}
}
