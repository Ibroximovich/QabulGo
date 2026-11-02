package repositories

import (
	"context"
	"errors"
	"time"

	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrLinkingTokenNotFound = errors.New("ulanish tokeni topilmadi yoki muddati o'tgan")
	ErrRecipientNotFound    = errors.New("telegram qabul qiluvchi topilmadi")
)

type TelegramRepository interface {
	CreateLinkingToken(ctx context.Context, token *models.TelegramLinkingToken) error
	GetValidLinkingTokenByHash(ctx context.Context, tokenHash string) (*models.TelegramLinkingToken, error)
	MarkLinkingTokenUsed(ctx context.Context, tokenID uuid.UUID) error

	UpsertClinicRecipient(ctx context.Context, recipient *models.ClinicTelegramRecipient) error
	GetClinicRecipients(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicTelegramRecipient, error)
	GetActiveClinicRecipients(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicTelegramRecipient, error)
	GetRecipientByID(ctx context.Context, clinicID, recipientID uuid.UUID) (*models.ClinicTelegramRecipient, error)
	DisconnectRecipient(ctx context.Context, clinicID, recipientID uuid.UUID) error
	ToggleRecipientNotifications(ctx context.Context, clinicID, recipientID uuid.UUID) (*models.ClinicTelegramRecipient, error)
	GetClinicsByTelegramUserID(ctx context.Context, telegramUserID int64) ([]models.Clinic, error)
	DeactivateRecipientByChatID(ctx context.Context, chatID int64) error

	IsUpdateProcessed(ctx context.Context, botType string, updateID int64) (bool, error)
	RecordUpdateProcessed(ctx context.Context, botType string, updateID int64) error

	GetAdminSession(ctx context.Context, telegramUserID int64) (*models.TelegramAdminSession, error)
	SetAdminSession(ctx context.Context, session *models.TelegramAdminSession) error
	ClearAdminSession(ctx context.Context, telegramUserID int64) error
}

type telegramRepository struct {
	db *gorm.DB
}

func NewTelegramRepository(db *gorm.DB) TelegramRepository {
	return &telegramRepository{db: db}
}

func (r *telegramRepository) CreateLinkingToken(ctx context.Context, token *models.TelegramLinkingToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *telegramRepository) GetValidLinkingTokenByHash(ctx context.Context, tokenHash string) (*models.TelegramLinkingToken, error) {
	var token models.TelegramLinkingToken
	nowUTC := time.Now().UTC()
	err := r.db.WithContext(ctx).
		Preload("Clinic").
		Where("token_hash = ? AND is_used = false AND expires_at > ?", tokenHash, nowUTC).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLinkingTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

func (r *telegramRepository) MarkLinkingTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.TelegramLinkingToken{}).
		Where("id = ?", tokenID).
		Updates(map[string]any{
			"is_used": true,
			"used_at": now,
		}).Error
}

func (r *telegramRepository) UpsertClinicRecipient(ctx context.Context, recipient *models.ClinicTelegramRecipient) error {
	recipient.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}, {Name: "telegram_chat_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"user_id",
				"telegram_user_id",
				"telegram_username",
				"telegram_first_name",
				"is_active",
				"notifications_enabled",
				"updated_at",
			}),
		}).
		Create(recipient).Error
}

func (r *telegramRepository) GetClinicRecipients(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicTelegramRecipient, error) {
	var list []models.ClinicTelegramRecipient
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *telegramRepository) GetActiveClinicRecipients(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicTelegramRecipient, error) {
	var list []models.ClinicTelegramRecipient
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND is_active = true AND notifications_enabled = true", clinicID).
		Find(&list).Error
	return list, err
}

func (r *telegramRepository) GetRecipientByID(ctx context.Context, clinicID, recipientID uuid.UUID) (*models.ClinicTelegramRecipient, error) {
	var recipient models.ClinicTelegramRecipient
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, recipientID).
		First(&recipient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecipientNotFound
		}
		return nil, err
	}
	return &recipient, nil
}

func (r *telegramRepository) DisconnectRecipient(ctx context.Context, clinicID, recipientID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.ClinicTelegramRecipient{}).
		Where("clinic_id = ? AND id = ?", clinicID, recipientID).
		Updates(map[string]any{
			"is_active":             false,
			"notifications_enabled": false,
			"updated_at":            time.Now(),
		}).Error
}

func (r *telegramRepository) ToggleRecipientNotifications(ctx context.Context, clinicID, recipientID uuid.UUID) (*models.ClinicTelegramRecipient, error) {
	recipient, err := r.GetRecipientByID(ctx, clinicID, recipientID)
	if err != nil {
		return nil, err
	}

	newStatus := !recipient.NotificationsEnabled
	err = r.db.WithContext(ctx).
		Model(recipient).
		Updates(map[string]any{
			"notifications_enabled": newStatus,
			"updated_at":            time.Now(),
		}).Error
	if err != nil {
		return nil, err
	}
	recipient.NotificationsEnabled = newStatus
	return recipient, nil
}

func (r *telegramRepository) GetClinicsByTelegramUserID(ctx context.Context, telegramUserID int64) ([]models.Clinic, error) {
	var clinics []models.Clinic
	err := r.db.WithContext(ctx).
		Table("clinics").
		Joins("JOIN clinic_telegram_recipients ctr ON ctr.clinic_id = clinics.id").
		Where("(ctr.telegram_user_id = ? OR ctr.telegram_chat_id = ?) AND ctr.is_active = true", telegramUserID, telegramUserID).
		Distinct().
		Find(&clinics).Error
	return clinics, err
}

func (r *telegramRepository) DeactivateRecipientByChatID(ctx context.Context, chatID int64) error {
	return r.db.WithContext(ctx).
		Model(&models.ClinicTelegramRecipient{}).
		Where("telegram_chat_id = ?", chatID).
		Updates(map[string]any{
			"is_active":             false,
			"notifications_enabled": false,
			"updated_at":            time.Now(),
		}).Error
}

func (r *telegramRepository) IsUpdateProcessed(ctx context.Context, botType string, updateID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TelegramUpdateDedup{}).
		Where("bot_type = ? AND update_id = ?", botType, updateID).
		Count(&count).Error
	return count > 0, err
}

func (r *telegramRepository) RecordUpdateProcessed(ctx context.Context, botType string, updateID int64) error {
	dedup := &models.TelegramUpdateDedup{
		BotType:     botType,
		UpdateID:    updateID,
		ProcessedAt: time.Now(),
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(dedup).Error
}

func (r *telegramRepository) GetAdminSession(ctx context.Context, telegramUserID int64) (*models.TelegramAdminSession, error) {
	var session models.TelegramAdminSession
	err := r.db.WithContext(ctx).
		Where("telegram_user_id = ?", telegramUserID).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *telegramRepository) SetAdminSession(ctx context.Context, session *models.TelegramAdminSession) error {
	session.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "telegram_user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"awaiting_action", "target_clinic_id", "updated_at"}),
		}).
		Create(session).Error
}

func (r *telegramRepository) ClearAdminSession(ctx context.Context, telegramUserID int64) error {
	return r.db.WithContext(ctx).
		Where("telegram_user_id = ?", telegramUserID).
		Delete(&models.TelegramAdminSession{}).Error
}

