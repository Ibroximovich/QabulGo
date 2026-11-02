package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/telegram"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTelegramNotConfigured = errors.New("telegram bot sozlamalari kiritilmagan")
	ErrUnauthorizedAdmin     = errors.New("ushbu telegram foydalanuvchisi superadmin huquqiga ega emas")
	ErrInvalidRejectionReason = errors.New("rad etish sababi kamida 5 ta belgidan iborat bo‘lishi shart")
)

type TelegramService interface {
	// Clinic Web Cabinet Endpoints
	GenerateClinicLinkingToken(ctx context.Context, clinicID, userID uuid.UUID) (*dto.TelegramLinkTokenResponse, error)
	GetClinicTelegramStatus(ctx context.Context, clinicID, userID uuid.UUID) (*dto.ClinicTelegramStatusResponse, error)
	DisconnectClinicRecipient(ctx context.Context, clinicID, userID, recipientID uuid.UUID) error
	ToggleClinicNotifications(ctx context.Context, clinicID, userID, recipientID uuid.UUID) (*dto.ClinicTelegramRecipientDTO, error)

	// Admin Cabinet Endpoints
	GetAdminTelegramStatus(ctx context.Context) (*dto.AdminTelegramStatusResponse, error)

	// Bot Update Processors
	ProcessClinicBotUpdate(ctx context.Context, update *telegram.Update) error
	ProcessAdminBotUpdate(ctx context.Context, update *telegram.Update) error

	// Outbox Message Dispatcher
	DispatchOutboxNotification(ctx context.Context, item *models.NotificationOutbox) error

	// Background Polling Listeners
	StartPolling(ctx context.Context)
}

type telegramService struct {
	cfg             *config.Config
	db              *gorm.DB
	telegramRepo    repositories.TelegramRepository
	outboxRepo      repositories.OutboxRepository
	clinicRepo      repositories.ClinicRepository
	appointmentRepo repositories.AppointmentRepository
	userRepo        repositories.UserRepository
	adminService    AdminService

	clinicBotClient telegram.Client
	adminBotClient  telegram.Client
}

func NewTelegramService(
	cfg *config.Config,
	db *gorm.DB,
	telegramRepo repositories.TelegramRepository,
	outboxRepo repositories.OutboxRepository,
	clinicRepo repositories.ClinicRepository,
	appointmentRepo repositories.AppointmentRepository,
	userRepo repositories.UserRepository,
	adminService AdminService,
) TelegramService {
	var clinicClient telegram.Client
	if cfg.TelegramClinicBotToken != "" {
		clinicClient = telegram.NewClient(cfg.TelegramClinicBotToken)
	}

	var adminClient telegram.Client
	if cfg.TelegramAdminBotToken != "" {
		adminClient = telegram.NewClient(cfg.TelegramAdminBotToken)
	}

	return &telegramService{
		cfg:             cfg,
		db:              db,
		telegramRepo:    telegramRepo,
		outboxRepo:      outboxRepo,
		clinicRepo:      clinicRepo,
		appointmentRepo: appointmentRepo,
		userRepo:        userRepo,
		adminService:    adminService,
		clinicBotClient: clinicClient,
		adminBotClient:  adminClient,
	}
}

// ----------------------------------------------------------------------------
// 1. Clinic Web Cabinet Linking Flow
// ----------------------------------------------------------------------------

func (s *telegramService) GenerateClinicLinkingToken(ctx context.Context, clinicID, userID uuid.UUID) (*dto.TelegramLinkTokenResponse, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	rawBytes := make([]byte, 24)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("tasodifiy token yaratishda xatolik: %w", err)
	}
	rawToken := hex.EncodeToString(rawBytes)

	hasher := sha256.New()
	hasher.Write([]byte(rawToken))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	tokenRecord := &models.TelegramLinkingToken{
		ID:        uuid.New(),
		ClinicID:  clinicID,
		UserID:    userID,
		TokenHash: tokenHash,
		IsUsed:    false,
		ExpiresAt: expiresAt,
	}

	if err := s.telegramRepo.CreateLinkingToken(ctx, tokenRecord); err != nil {
		return nil, fmt.Errorf("tokenni saqlashda xatolik: %w", err)
	}

	botUsername := strings.TrimPrefix(strings.TrimSpace(s.cfg.TelegramClinicBotUsername), "@")
	if botUsername == "" {
		botUsername = "qabulgo_clinic_bot"
	}
	deepLink := fmt.Sprintf("https://t.me/%s?start=%s", botUsername, rawToken)

	return &dto.TelegramLinkTokenResponse{
		Token:       rawToken,
		DeepLink:    deepLink,
		BotUsername: botUsername,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *telegramService) GetClinicTelegramStatus(ctx context.Context, clinicID, userID uuid.UUID) (*dto.ClinicTelegramStatusResponse, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	recipients, err := s.telegramRepo.GetClinicRecipients(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	recipientDTOs := make([]dto.ClinicTelegramRecipientDTO, 0, len(recipients))
	hasActive := false
	for _, r := range recipients {
		if r.IsActive {
			hasActive = true
		}
		recipientDTOs = append(recipientDTOs, dto.ClinicTelegramRecipientDTO{
			ID:                   r.ID,
			TelegramUserID:       r.TelegramUserID,
			TelegramUsername:     r.TelegramUsername,
			TelegramFirstName:    r.TelegramFirstName,
			IsActive:             r.IsActive,
			NotificationsEnabled: r.NotificationsEnabled,
			CreatedAt:            r.CreatedAt,
		})
	}

	return &dto.ClinicTelegramStatusResponse{
		IsConnected: hasActive,
		BotUsername: s.cfg.TelegramClinicBotUsername,
		Recipients:  recipientDTOs,
	}, nil
}

func (s *telegramService) DisconnectClinicRecipient(ctx context.Context, clinicID, userID, recipientID uuid.UUID) error {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return repositories.ErrClinicAccessDenied
	}

	return s.telegramRepo.DisconnectRecipient(ctx, clinicID, recipientID)
}

func (s *telegramService) ToggleClinicNotifications(ctx context.Context, clinicID, userID, recipientID uuid.UUID) (*dto.ClinicTelegramRecipientDTO, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	updated, err := s.telegramRepo.ToggleRecipientNotifications(ctx, clinicID, recipientID)
	if err != nil {
		return nil, err
	}

	return &dto.ClinicTelegramRecipientDTO{
		ID:                   updated.ID,
		TelegramUserID:       updated.TelegramUserID,
		TelegramUsername:     updated.TelegramUsername,
		TelegramFirstName:    updated.TelegramFirstName,
		IsActive:             updated.IsActive,
		NotificationsEnabled: updated.NotificationsEnabled,
		CreatedAt:            updated.CreatedAt,
	}, nil
}

func (s *telegramService) GetAdminTelegramStatus(ctx context.Context) (*dto.AdminTelegramStatusResponse, error) {
	botUsername := strings.TrimPrefix(strings.TrimSpace(s.cfg.TelegramAdminBotUsername), "@")
	if botUsername == "" {
		botUsername = "qabulgo_admin_bot"
	}

	deepLink := fmt.Sprintf("https://t.me/%s", botUsername)
	isConfigured := s.cfg.TelegramAdminBotToken != ""

	return &dto.AdminTelegramStatusResponse{
		IsConfigured:    isConfigured,
		BotUsername:     botUsername,
		BotName:         "QabulGo Admin Bot",
		DeepLink:        deepLink,
		AllowedAdminIDs: s.cfg.TelegramAdminAllowedIDs,
	}, nil
}

// ----------------------------------------------------------------------------
// 2. Clinic Bot Logic & Update Handlers
// ----------------------------------------------------------------------------

func (s *telegramService) ProcessClinicBotUpdate(ctx context.Context, update *telegram.Update) error {
	if s.clinicBotClient == nil {
		return nil
	}

	// Update Deduplication
	processed, err := s.telegramRepo.IsUpdateProcessed(ctx, "clinic", update.UpdateID)
	if err == nil && processed {
		return nil
	}
	_ = s.telegramRepo.RecordUpdateProcessed(ctx, "clinic", update.UpdateID)

	if update.Message != nil {
		return s.handleClinicMessage(ctx, update.Message)
	}
	if update.CallbackQuery != nil {
		return s.handleClinicCallbackQuery(ctx, update.CallbackQuery)
	}
	return nil
}

func (s *telegramService) handleClinicMessage(ctx context.Context, msg *telegram.Message) error {
	// Only private chat allowed for privacy
	if msg.Chat.Type != "private" {
		return nil
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil
	}
	chatID := msg.Chat.ID
	fromUser := msg.From
	if fromUser == nil {
		return nil
	}

	// Split message into command and arguments
	fields := strings.Fields(text)
	rawCmd := fields[0]
	cmd := strings.ToLower(rawCmd)
	if atIdx := strings.Index(cmd, "@"); atIdx != -1 {
		cmd = cmd[:atIdx]
	}

	var arg string
	if len(fields) > 1 {
		arg = strings.TrimSpace(fields[1])
	}

	// 1. Deep Linking Token (/start <token> or /start@bot <token>)
	if cmd == "/start" && arg != "" {
		return s.completeClinicLinking(ctx, chatID, fromUser, arg)
	}

	// 2. Plain /start or interactive commands
	if cmd == "/start" {
		clinics, err := s.telegramRepo.GetClinicsByTelegramUserID(ctx, fromUser.ID)
		if err != nil || len(clinics) == 0 {
			replyText := fmt.Sprintf("👋 <b>Assalomu alaykum, %s!</b>\n\n"+
				"Bu <b>QabulGo Clinic Bot</b> — klinika xodimlari uchun xabarnomalar boti.\n\n"+
				"⚠️ Sizning Telegram akkauntingiz hali hech qaysi klinikaga ulanmagan.\n\n"+
				"<b>Ulanish uchun:</b>\n"+
				"1. QabulGo klinika kabinetingizga kiring\n"+
				"2. Chap menyudan <b>Telegram bot</b> bo‘limini oching\n"+
				"3. <b>«Telegram botni ulash»</b> tugmasini bosing\n\n"+
				"Savollar bo‘lsa, /help buyrug‘ini yuboring.", fromUser.FirstName)
			_, sendErr := s.clinicBotClient.SendMessage(ctx, chatID, replyText, nil)
			if sendErr != nil {
				log.Printf("[Telegram Clinic Bot] SendMessage xatoligi (chatID=%d): %v\n", chatID, sendErr)
			}
			return sendErr
		}

		clinicNames := make([]string, 0, len(clinics))
		for _, c := range clinics {
			clinicNames = append(clinicNames, "• "+c.Name)
		}

		replyText := fmt.Sprintf("👋 <b>Assalomu alaykum, %s!</b>\n\nSiz quyidagi klinikalar xabarnomasiga ulangansiz:\n%s\n\nKerakli bo‘limni tanlang:",
			fromUser.FirstName, strings.Join(clinicNames, "\n"))

		markup := s.buildClinicMenuMarkup()
		_, sendErr := s.clinicBotClient.SendMessage(ctx, chatID, replyText, markup)
		if sendErr != nil {
			log.Printf("[Telegram Clinic Bot] SendMessage xatoligi (chatID=%d): %v\n", chatID, sendErr)
		}
		return sendErr
	}

	switch {
	case cmd == "/today" || text == "📅 Bugungi bronlar":
		return s.sendClinicAppointmentsForDate(ctx, chatID, fromUser.ID, s.getTashkentDateOffset(0), "Bugungi")

	case cmd == "/tomorrow" || text == "🗓 Ertangi bronlar":
		return s.sendClinicAppointmentsForDate(ctx, chatID, fromUser.ID, s.getTashkentDateOffset(1), "Ertangi")

	case cmd == "/appointments" || text == "📋 Kelgusi bronlar":
		return s.sendClinicUpcomingAppointments(ctx, chatID, fromUser.ID, 0)

	case cmd == "/help" || text == "ℹ️ Yordam":
		helpText := "<b>QabulGo Clinic Bot buyruqlari:</b>\n\n" +
			"/today — Bugungi kun qabullari\n" +
			"/tomorrow — Ertangi kun qabullari\n" +
			"/appointments — Kelgusi barcha tasdiqlangan qabullar\n" +
			"/start — Ulanish holati va asosiy menyu\n" +
			"/help — Yordam ma’lumotlari\n\n" +
			"Savollar bo‘lsa, sayt orqali texnik ko‘makka murojaat qiling."
		_, sendErr := s.clinicBotClient.SendMessage(ctx, chatID, helpText, s.buildClinicMenuMarkup())
		return sendErr

	default:
		// Not recognized command - provide friendly fallback instead of silence
		clinics, _ := s.telegramRepo.GetClinicsByTelegramUserID(ctx, fromUser.ID)
		if len(clinics) > 0 {
			replyText := "ℹ️ Iltimos, quyidagi menyu buyruqlaridan foydalaning yoki /help deb yozing:"
			_, _ = s.clinicBotClient.SendMessage(ctx, chatID, replyText, s.buildClinicMenuMarkup())
		} else {
			replyText := "👋 <b>QabulGo Clinic Bot</b>\n\n" +
				"⚠️ Akkauntingiz klinikaga ulanmagan. Iltimos, klinika kabinetingizdan <b>Telegram bot</b> bo‘limiga kirib ulaning."
			_, _ = s.clinicBotClient.SendMessage(ctx, chatID, replyText, nil)
		}
		return nil
	}
}

func (s *telegramService) completeClinicLinking(ctx context.Context, chatID int64, fromUser *telegram.User, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		replyText := "❌ <b>Ulanish kodi topilmadi.</b>\n\nIltimos, QabulGo klinika kabinetidagi «Telegram botni ulash» tugmasi orqali qaytadan urinib ko‘ring."
		_, _ = s.clinicBotClient.SendMessage(ctx, chatID, replyText, nil)
		return nil
	}

	hasher := sha256.New()
	hasher.Write([]byte(rawToken))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	tokenRecord, err := s.telegramRepo.GetValidLinkingTokenByHash(ctx, tokenHash)
	if err != nil {
		log.Printf("[Telegram Clinic Bot] Linking token topilmadi yoki muddati o'tgan (hash: %s): %v\n", tokenHash, err)
		replyText := "❌ <b>Ulanish kodi yaroqsiz yoki muddati o‘tgan.</b>\n\nIltimos, QabulGo klinika kabinetidan yangi havola oling."
		_, _ = s.clinicBotClient.SendMessage(ctx, chatID, replyText, nil)
		return nil
	}

	// Mark token used
	_ = s.telegramRepo.MarkLinkingTokenUsed(ctx, tokenRecord.ID)

	// Upsert recipient
	recipient := &models.ClinicTelegramRecipient{
		ID:                   uuid.New(),
		ClinicID:             tokenRecord.ClinicID,
		UserID:               &tokenRecord.UserID,
		TelegramUserID:       fromUser.ID,
		TelegramChatID:       chatID,
		TelegramUsername:     fromUser.Username,
		TelegramFirstName:    fromUser.FirstName,
		IsActive:             true,
		NotificationsEnabled: true,
	}

	if err := s.telegramRepo.UpsertClinicRecipient(ctx, recipient); err != nil {
		log.Printf("Recipientni saqlashda xatolik: %v\n", err)
		_, _ = s.clinicBotClient.SendMessage(ctx, chatID, "❌ Ulanishda texnik xatolik yuz berdi. Iltimos qaytadan urinib ko‘ring.", nil)
		return err
	}

	clinicName := "Klinika"
	if tokenRecord.Clinic != nil && tokenRecord.Clinic.Name != "" {
		clinicName = tokenRecord.Clinic.Name
	}

	welcomeText := fmt.Sprintf("✅ <b>Tabriklaymiz!</b>\n\nSizning Telegram akkauntingiz <b>%s</b> klinikasiga muvaffaqiyatli bog‘landi.\n\n"+
		"🔔 Endi bemorlar qabulga yozilganda yoki qabul bekor qilinganda ushbu bot orqali darhol xabardor qilinasiz!\n\n"+
		"Quyidagi menyu orqali qabullarni ko‘rishingiz mumkin:", clinicName)

	markup := s.buildClinicMenuMarkup()
	_, _ = s.clinicBotClient.SendMessage(ctx, chatID, welcomeText, markup)
	return nil
}

func (s *telegramService) handleClinicCallbackQuery(ctx context.Context, cb *telegram.CallbackQuery) error {
	data := cb.Data
	chatID := cb.From.ID
	_ = s.clinicBotClient.AnswerCallbackQuery(ctx, cb.ID, "", false)

	if data == "cb:clinic_today" {
		return s.sendClinicAppointmentsForDate(ctx, chatID, cb.From.ID, s.getTashkentDateOffset(0), "Bugungi")
	}
	if data == "cb:clinic_tomorrow" {
		return s.sendClinicAppointmentsForDate(ctx, chatID, cb.From.ID, s.getTashkentDateOffset(1), "Ertangi")
	}
	if strings.HasPrefix(data, "cb:clinic_upcoming:") {
		offsetStr := strings.TrimPrefix(data, "cb:clinic_upcoming:")
		offset, _ := strconv.Atoi(offsetStr)
		return s.sendClinicUpcomingAppointments(ctx, chatID, cb.From.ID, offset)
	}

	return nil
}

func (s *telegramService) sendClinicAppointmentsForDate(ctx context.Context, chatID int64, tgUserID int64, dateStr, dateLabel string) error {
	clinics, err := s.telegramRepo.GetClinicsByTelegramUserID(ctx, tgUserID)
	if err != nil || len(clinics) == 0 {
		_, _ = s.clinicBotClient.SendMessage(ctx, chatID, "⚠️ Siz ulangan klinika topilmadi.", nil)
		return nil
	}

	for _, c := range clinics {
		filter := &dto.ClinicAppointmentFilter{
			Date:   dateStr,
			Status: "confirmed",
			Page:   1,
			Limit:  20,
		}
		res, _, err := s.appointmentRepo.ListForClinic(ctx, c.ID, filter)
		if err != nil {
			continue
		}

		if len(res) == 0 {
			text := fmt.Sprintf("📅 <b>%s (%s) bo‘yicha qabullar:</b>\n🏥 <b>%s</b>\n\n<i>Ushbu kunga tasdiqlangan qabullar mavjud emas.</i>",
				dateLabel, dateStr, c.Name)
			_, _ = s.clinicBotClient.SendMessage(ctx, chatID, text, nil)
			continue
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📅 <b>%s (%s) qabullari:</b>\n🏥 <b>%s</b>\n\n", dateLabel, dateStr, c.Name))

		for idx, a := range res {
			patientName := "Bemor"
			patientPhone := "Mavjud emas"
			if a.Patient != nil {
				patientName = fmt.Sprintf("%s %s", a.Patient.FirstName, a.Patient.LastName)
				patientPhone = a.Patient.Phone
			}
			docName := "Shifokor"
			if a.Doctor != nil {
				docName = fmt.Sprintf("%s %s", a.Doctor.FirstName, a.Doctor.LastName)
			}

			sb.WriteString(fmt.Sprintf("<b>%d. %s–%s</b>\n👤 Bemor: %s\n📞 Telefon: %s\n👨‍⚕️ Shifokor: %s\n✅ Holat: Tasdiqlangan\n\n",
				idx+1, a.StartTime, a.EndTime, patientName, patientPhone, docName))
		}

		markup := &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text: "🌐 Saytda to‘liq ko‘rish",
						URL:  s.buildWebURL("/clinic/appointments"),
					},
				},
			},
		}

		_, _ = s.clinicBotClient.SendMessage(ctx, chatID, sb.String(), markup)
	}

	return nil
}

func (s *telegramService) sendClinicUpcomingAppointments(ctx context.Context, chatID int64, tgUserID int64, offset int) error {
	clinics, err := s.telegramRepo.GetClinicsByTelegramUserID(ctx, tgUserID)
	if err != nil || len(clinics) == 0 {
		_, _ = s.clinicBotClient.SendMessage(ctx, chatID, "⚠️ Siz ulangan klinika topilmadi.", nil)
		return nil
	}

	todayStr := s.getTashkentDateOffset(0)
	limit := 5

	for _, c := range clinics {
		page := (offset / limit) + 1
		filter := &dto.ClinicAppointmentFilter{
			DateFrom: todayStr,
			Status:   "confirmed",
			Page:     page,
			Limit:    limit,
		}
		res, _, err := s.appointmentRepo.ListForClinic(ctx, c.ID, filter)
		if err != nil {
			continue
		}

		if len(res) == 0 && offset == 0 {
			text := fmt.Sprintf("📋 <b>Kelgusi qabullar:</b>\n🏥 <b>%s</b>\n\n<i>Kelgusida tasdiqlangan qabullar mavjud emas.</i>", c.Name)
			_, _ = s.clinicBotClient.SendMessage(ctx, chatID, text, nil)
			continue
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📋 <b>Kelgusi qabullar (%d–%d):</b>\n🏥 <b>%s</b>\n\n", offset+1, offset+len(res), c.Name))

		for idx, a := range res {
			patientName := "Bemor"
			patientPhone := "Mavjud emas"
			if a.Patient != nil {
				patientName = fmt.Sprintf("%s %s", a.Patient.FirstName, a.Patient.LastName)
				patientPhone = a.Patient.Phone
			}
			docName := "Shifokor"
			if a.Doctor != nil {
				docName = fmt.Sprintf("%s %s", a.Doctor.FirstName, a.Doctor.LastName)
			}

			sb.WriteString(fmt.Sprintf("<b>%d. %s (%s–%s)</b>\n👤 Bemor: %s\n📞 Tel: %s\n👨‍⚕️ Shifokor: %s\n\n",
				offset+idx+1, a.AppointmentDate, a.StartTime, a.EndTime, patientName, patientPhone, docName))
		}

		var navButtons []telegram.InlineKeyboardButton
		if offset > 0 {
			prevOffset := offset - limit
			if prevOffset < 0 {
				prevOffset = 0
			}
			navButtons = append(navButtons, telegram.InlineKeyboardButton{
				Text:         "⬅️ Oldingi",
				CallbackData: fmt.Sprintf("cb:clinic_upcoming:%d", prevOffset),
			})
		}
		if len(res) == limit {
			navButtons = append(navButtons, telegram.InlineKeyboardButton{
				Text:         "Keyingi ➡️",
				CallbackData: fmt.Sprintf("cb:clinic_upcoming:%d", offset+limit),
			})
		}

		rows := [][]telegram.InlineKeyboardButton{}
		if len(navButtons) > 0 {
			rows = append(rows, navButtons)
		}
		rows = append(rows, []telegram.InlineKeyboardButton{
			{
				Text: "🌐 Saytda ko‘rish",
				URL:  s.buildWebURL("/clinic/appointments"),
			},
		})

		markup := &telegram.InlineKeyboardMarkup{InlineKeyboard: rows}
		_, _ = s.clinicBotClient.SendMessage(ctx, chatID, sb.String(), markup)
	}

	return nil
}

func (s *telegramService) buildClinicMenuMarkup() *telegram.InlineKeyboardMarkup {
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "📅 Bugungi bronlar", CallbackData: "cb:clinic_today"},
				{Text: "🗓 Ertangi bronlar", CallbackData: "cb:clinic_tomorrow"},
			},
			{
				{Text: "📋 Kelgusi bronlar", CallbackData: "cb:clinic_upcoming:0"},
				{Text: "🌐 Saytga o‘tish", URL: s.buildWebURL("/clinic/appointments")},
			},
		},
	}
}

// ----------------------------------------------------------------------------
// 3. Admin Bot Logic & Approval Workflows
// ----------------------------------------------------------------------------

func (s *telegramService) ProcessAdminBotUpdate(ctx context.Context, update *telegram.Update) error {
	if s.adminBotClient == nil {
		return nil
	}

	// Update Deduplication
	processed, err := s.telegramRepo.IsUpdateProcessed(ctx, "admin", update.UpdateID)
	if err == nil && processed {
		return nil
	}
	_ = s.telegramRepo.RecordUpdateProcessed(ctx, "admin", update.UpdateID)

	if update.Message != nil {
		return s.handleAdminMessage(ctx, update.Message)
	}
	if update.CallbackQuery != nil {
		return s.handleAdminCallbackQuery(ctx, update.CallbackQuery)
	}
	return nil
}

func (s *telegramService) isAuthorizedAdmin(telegramUserID int64) bool {
	if len(s.cfg.TelegramAdminAllowedIDs) == 0 {
		return false
	}
	for _, id := range s.cfg.TelegramAdminAllowedIDs {
		if id == telegramUserID {
			return true
		}
	}
	return false
}

func (s *telegramService) handleAdminMessage(ctx context.Context, msg *telegram.Message) error {
	if msg.Chat.Type != "private" {
		return nil
	}

	chatID := msg.Chat.ID
	fromUser := msg.From
	if fromUser == nil {
		return nil
	}

	// Authorization Check
	if !s.isAuthorizedAdmin(fromUser.ID) {
		text := fmt.Sprintf("⛔️ <b>Ruxsat etilmagan!</b>\n\nSizda QabulGo Nazorat Markazi (Superadmin) botidan foydalanish huquqi yo‘q.\n\nSizning Telegram ID raqamingiz: <code>%d</code>\nUshbu ID'ni backend .env faylidagi <code>%s</code> sozlamasiga qo‘shing.", fromUser.ID, "TELEGRAM_ADMIN_ALLOWED_IDS")
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, text, nil)
		return nil
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil
	}

	fields := strings.Fields(text)
	cmd := strings.ToLower(fields[0])
	if atIdx := strings.Index(cmd, "@"); atIdx != -1 {
		cmd = cmd[:atIdx]
	}

	// Check if in active interactive session (e.g. awaiting rejection reason)
	session, _ := s.telegramRepo.GetAdminSession(ctx, fromUser.ID)
	if session != nil && session.AwaitingAction == "reject_clinic" && session.TargetClinicID != nil {
		if cmd == "/cancel" || text == "/cancel" {
			_ = s.telegramRepo.ClearAdminSession(ctx, fromUser.ID)
			_, _ = s.adminBotClient.SendMessage(ctx, chatID, "Amal bekor qilindi.", s.buildAdminMenuMarkup())
			return nil
		}

		if len(text) < 5 {
			_, _ = s.adminBotClient.SendMessage(ctx, chatID,
				"⚠️ Rad etish sababi kamida 5 ta belgidan iborat bo‘lishi kerak. Iltimos qaytadan yozing yoki /cancel buyrug‘ini bering:", nil)
			return nil
		}

		clinicID := *session.TargetClinicID
		_ = s.telegramRepo.ClearAdminSession(ctx, fromUser.ID)

		// Execute rejection through AdminService
		adminUUID := s.getSystemSuperAdminUUID(ctx)
		err := s.adminService.RejectClinic(ctx, adminUUID, clinicID, text, "telegram_bot", "TelegramAdminBot/1.0")
		if err != nil {
			_, _ = s.adminBotClient.SendMessage(ctx, chatID, fmt.Sprintf("❌ Klinikani rad etishda xatolik: %v", err), nil)
			return nil
		}

		clinic, _ := s.clinicRepo.GetByID(ctx, clinicID)
		clinicName := "Klinika"
		if clinic != nil {
			clinicName = clinic.Name
		}

		replyText := fmt.Sprintf("❌ <b>KLINIKA RAD ETILDI</b>\n\n🏥 <b>%s</b>\n📝 <b>Sabab:</b> %s\n👤 <b>Mas’ul:</b> Superadmin (@%s)\n📅 <b>Vaqt:</b> %s",
			clinicName, text, fromUser.Username, time.Now().Format("02.01.2006 15:04"))

		markup := &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text: "🌐 Saytda ko‘rish",
						URL:  s.buildWebURL(fmt.Sprintf("/admin/clinics?clinicId=%s", clinicID.String())),
					},
				},
			},
		}

		_, _ = s.adminBotClient.SendMessage(ctx, chatID, replyText, markup)
		return nil
	}

	switch {
	case cmd == "/start":
		welcomeText := fmt.Sprintf("🛡 <b>QabulGo Nazorat Markazi (Superadmin Bot)</b>\n\nAssalomu alaykum, <b>%s</b>!\n\n"+
			"Ushbu bot orqali siz yangi klinika arizalarini ko‘rib chiqishingiz, tasdiqlashingiz yoki rad etishingiz mumkin.", fromUser.FirstName)
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, welcomeText, s.buildAdminMenuMarkup())
		return nil

	case cmd == "/pending" || text == "⏳ Kutilayotgan klinikalar":
		return s.sendAdminPendingClinics(ctx, chatID, 0)

	case cmd == "/approved" || text == "✅ Tasdiqlangan klinikalar":
		return s.sendAdminApprovedClinics(ctx, chatID)

	case cmd == "/help" || text == "ℹ️ Yordam":
		helpText := "<b>Superadmin Bot buyruqlari:</b>\n\n" +
			"/pending — Tasdiqlash kutilayotgan klinikalar arizalari\n" +
			"/approved — Yaqinda tasdiqlangan klinikalar\n" +
			"/start — Asosiy menyu\n" +
			"/cancel — Joriy amalni bekor qilish"
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, helpText, s.buildAdminMenuMarkup())
		return nil

	default:
		replyText := "ℹ️ Iltimos, quyidagi menyu buyruqlaridan foydalaning yoki /help deb yozing:"
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, replyText, s.buildAdminMenuMarkup())
		return nil
	}

	return nil
}

func (s *telegramService) handleAdminCallbackQuery(ctx context.Context, cb *telegram.CallbackQuery) error {
	fromUser := cb.From
	chatID := fromUser.ID

	if !s.isAuthorizedAdmin(fromUser.ID) {
		_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Sizda superadmin ruxsati yo‘q", true)
		return nil
	}

	data := cb.Data

	// 1. Approve Clinic
	if strings.HasPrefix(data, "cb:admin_approve:") {
		clinicIDStr := strings.TrimPrefix(data, "cb:admin_approve:")
		clinicID, err := uuid.Parse(clinicIDStr)
		if err != nil {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Noto‘g‘ri klinika ID", true)
			return nil
		}

		clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
		if err != nil {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Klinika topilmadi", true)
			return nil
		}

		if clinic.Status == models.ClinicStatusApproved {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Bu klinika allaqachon tasdiqlangan", true)
			return nil
		}
		if clinic.Status != models.ClinicStatusPending {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, fmt.Sprintf("Klinika holati: %s", clinic.Status), true)
			return nil
		}

		adminUUID := s.getSystemSuperAdminUUID(ctx)
		err = s.adminService.ApproveClinic(ctx, adminUUID, clinicID, "telegram_bot", "TelegramAdminBot/1.0")
		if err != nil {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, fmt.Sprintf("Xatolik: %v", err), true)
			return nil
		}

		_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Klinika tasdiqlandi!", false)

		updatedText := fmt.Sprintf("✅ <b>KLINIKA TASDIQLANDI</b>\n\n🏥 <b>%s</b>\n👤 <b>Tasdiqlagan:</b> Superadmin (@%s)\n📅 <b>Sana va vaqt:</b> %s",
			clinic.Name, fromUser.Username, time.Now().Format("02.01.2006 15:04"))

		markup := &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text: "🌐 Saytda ko‘rish",
						URL:  s.buildWebURL(fmt.Sprintf("/admin/clinics?clinicId=%s", clinic.ID.String())),
					},
				},
			},
		}

		if cb.Message != nil {
			_, _ = s.adminBotClient.EditMessageText(ctx, chatID, cb.Message.MessageID, updatedText, markup)
		}
		return nil
	}

	// 2. Reject Clinic Prompt
	if strings.HasPrefix(data, "cb:admin_reject_prompt:") {
		clinicIDStr := strings.TrimPrefix(data, "cb:admin_reject_prompt:")
		clinicID, err := uuid.Parse(clinicIDStr)
		if err != nil {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Noto‘g‘ri klinika ID", true)
			return nil
		}

		clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
		if err != nil {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Klinika topilmadi", true)
			return nil
		}

		if clinic.Status != models.ClinicStatusPending {
			_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, fmt.Sprintf("Klinika kutilayotgan holatda emas (%s)", clinic.Status), true)
			return nil
		}

		// Save session state
		_ = s.telegramRepo.SetAdminSession(ctx, &models.TelegramAdminSession{
			ID:             uuid.New(),
			TelegramUserID: fromUser.ID,
			AwaitingAction: "reject_clinic",
			TargetClinicID: &clinicID,
		})

		_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "Sababni kiritish kutilmoqda", false)

		promptText := fmt.Sprintf("📝 <b>%s</b> klinikasini rad etish sababini yozib yuboring (kamida 5 ta belgi):\n\n<i>(Bekor qilish uchun /cancel deb yozing)</i>", clinic.Name)
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, promptText, nil)
		return nil
	}

	// 3. Pending List Pagination
	if strings.HasPrefix(data, "cb:admin_pending:") {
		offsetStr := strings.TrimPrefix(data, "cb:admin_pending:")
		offset, _ := strconv.Atoi(offsetStr)
		_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "", false)
		return s.sendAdminPendingClinics(ctx, chatID, offset)
	}

	// 4. Approved Clinics List
	if data == "cb:admin_approved" || strings.HasPrefix(data, "cb:admin_approved:") {
		_ = s.adminBotClient.AnswerCallbackQuery(ctx, cb.ID, "", false)
		return s.sendAdminApprovedClinics(ctx, chatID)
	}

	return nil
}

func (s *telegramService) sendAdminPendingClinics(ctx context.Context, chatID int64, offset int) error {
	limit := 3
	page := (offset / limit) + 1
	filter := &dto.AdminClinicFilter{
		Status: "pending",
		Page:   page,
		Limit:  limit,
	}

	clinics, total, err := s.adminService.ListClinics(ctx, filter)
	if err != nil {
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, fmt.Sprintf("Xatolik: %v", err), nil)
		return nil
	}

	if total == 0 {
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, "⏳ Hozirda tasdiqlash kutilayotgan yangi klinikalar yo‘q.", s.buildAdminMenuMarkup())
		return nil
	}

	for _, c := range clinics {
		text := fmt.Sprintf("🏥 <b>YANGI KLINIKA SO‘ROVI</b>\n\n"+
			"📌 <b>Klinika:</b> %s\n"+
			"📞 <b>Telefon:</b> %s\n"+
			"📧 <b>Email:</b> %s\n"+
			"📍 <b>Hudud:</b> %s, %s\n"+
			"🏢 <b>Manzil:</b> %s\n"+
			"📅 <b>Yuborilgan sana:</b> %s\n"+
			"⏳ <b>Holat:</b> Tasdiqlash kutilmoqda",
			c.Name, c.Phone, c.Email, c.Region, c.District, c.Address, c.CreatedAt.Format("02.01.2006 15:04"))

		markup := &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{Text: "✅ Tasdiqlash", CallbackData: fmt.Sprintf("cb:admin_approve:%s", c.ID.String())},
					{Text: "❌ Rad etish", CallbackData: fmt.Sprintf("cb:admin_reject_prompt:%s", c.ID.String())},
				},
				{
					{
						Text: "🌐 Saytda to‘liq ko‘rish",
						URL:  s.buildWebURL(fmt.Sprintf("/admin/clinics?clinicId=%s", c.ID.String())),
					},
				},
			},
		}

		_, _ = s.adminBotClient.SendMessage(ctx, chatID, text, markup)
	}

	// Pagination indicator
	var navButtons []telegram.InlineKeyboardButton
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		navButtons = append(navButtons, telegram.InlineKeyboardButton{
			Text:         "⬅️ Oldingi",
			CallbackData: fmt.Sprintf("cb:admin_pending:%d", prevOffset),
		})
	}
	if offset+limit < int(total) {
		navButtons = append(navButtons, telegram.InlineKeyboardButton{
			Text:         "Keyingi ➡️",
			CallbackData: fmt.Sprintf("cb:admin_pending:%d", offset+limit),
		})
	}

	if len(navButtons) > 0 {
		markup := &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{navButtons},
		}
		infoText := fmt.Sprintf("Jami kutilayotgan: %d ta (ko‘rsatilmoqda: %d–%d)", total, offset+1, offset+len(clinics))
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, infoText, markup)
	}

	return nil
}

func (s *telegramService) sendAdminApprovedClinics(ctx context.Context, chatID int64) error {
	filter := &dto.AdminClinicFilter{
		Status: "approved",
		Page:   1,
		Limit:  5,
	}

	clinics, total, err := s.adminService.ListClinics(ctx, filter)
	if err != nil {
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, fmt.Sprintf("Xatolik: %v", err), nil)
		return nil
	}

	if total == 0 {
		_, _ = s.adminBotClient.SendMessage(ctx, chatID, "Hozircha tasdiqlangan klinikalar yo‘q.", s.buildAdminMenuMarkup())
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ <b>Yaqinda tasdiqlangan klinikalar (jami: %d ta):</b>\n\n", total))
	for idx, c := range clinics {
		sb.WriteString(fmt.Sprintf("<b>%d. %s</b>\n📍 %s, %s\n📞 %s\n\n", idx+1, c.Name, c.Region, c.District, c.Phone))
	}

	markup := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "🌐 Admin panelga o‘tish", URL: s.buildWebURL("/admin/clinics")},
			},
		},
	}

	_, _ = s.adminBotClient.SendMessage(ctx, chatID, sb.String(), markup)
	return nil
}

func (s *telegramService) buildAdminMenuMarkup() *telegram.InlineKeyboardMarkup {
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "⏳ Kutilayotgan klinikalar", CallbackData: "cb:admin_pending:0"},
				{Text: "✅ Tasdiqlangan klinikalar", CallbackData: "cb:admin_approved"},
			},
			{
				{Text: "🌐 Admin panelga o‘tish", URL: s.buildWebURL("/admin/clinics")},
			},
		},
	}
}

func (s *telegramService) getSystemSuperAdminUUID(ctx context.Context) uuid.UUID {
	var user models.User
	err := s.db.WithContext(ctx).Where("role = 'super_admin'").First(&user).Error
	if err == nil {
		return user.ID
	}
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// ----------------------------------------------------------------------------
// 4. Outbox Dispatcher Implementation
// ----------------------------------------------------------------------------

func (s *telegramService) DispatchOutboxNotification(ctx context.Context, item *models.NotificationOutbox) error {
	switch item.EventType {
	case "appointment_created":
		return s.dispatchAppointmentCreated(ctx, item)
	case "appointment_cancelled":
		return s.dispatchAppointmentCancelled(ctx, item)
	case "clinic_submitted":
		return s.dispatchClinicSubmitted(ctx, item)
	default:
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Noma'lum event turi: "+item.EventType)
		return nil
	}
}

func (s *telegramService) dispatchAppointmentCreated(ctx context.Context, item *models.NotificationOutbox) error {
	if s.clinicBotClient == nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Clinic Bot sozlanmagan")
		return nil
	}

	if item.TargetClinicID == nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "TargetClinicID yo‘q")
		return nil
	}

	recipients, err := s.telegramRepo.GetActiveClinicRecipients(ctx, *item.TargetClinicID)
	if err != nil || len(recipients) == 0 {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Ulangan faol recipientlar yo‘q")
		return nil
	}

	var data struct {
		AppointmentID string  `json:"appointment_id"`
		ClinicName    string  `json:"clinic_name"`
		PatientName   string  `json:"patient_name"`
		PatientPhone  string  `json:"patient_phone"`
		DoctorName    string  `json:"doctor_name"`
		Specialty     string  `json:"specialty"`
		Date          string  `json:"date"`
		StartTime     string  `json:"start_time"`
		EndTime       string  `json:"end_time"`
		Price         float64 `json:"price"`
	}
	if err := json.Unmarshal([]byte(item.Payload), &data); err != nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Payload parslash xatoligi")
		return nil
	}

	formattedPrice := fmt.Sprintf("%.0f so‘m", data.Price)
	text := fmt.Sprintf("🔔 <b>YANGI QABUL!</b>\n\n"+
		"🏥 <b>Klinika:</b> %s\n"+
		"👤 <b>Bemor:</b> %s\n"+
		"📞 <b>Telefon:</b> %s\n"+
		"👨‍⚕️ <b>Shifokor:</b> %s\n"+
		"🩺 <b>Mutaxassislik:</b> %s\n"+
		"📅 <b>Sana:</b> %s\n"+
		"🕙 <b>Vaqt:</b> %s–%s\n"+
		"💰 <b>Qabul narxi:</b> %s\n"+
		"✅ <b>Holat:</b> Tasdiqlangan",
		data.ClinicName, data.PatientName, data.PatientPhone, data.DoctorName, data.Specialty,
		data.Date, data.StartTime, data.EndTime, formattedPrice)

	markup := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{
					Text: "🌐 Saytga o‘tish",
					URL:  s.buildWebURL(fmt.Sprintf("/clinic/appointments/%s", data.AppointmentID)),
				},
			},
		},
	}

	allSent := true
	for _, r := range recipients {
		_, err := s.clinicBotClient.SendMessage(ctx, r.TelegramChatID, text, markup)
		if err != nil {
			if errors.Is(err, telegram.ErrBotBlockedByUser) {
				_ = s.telegramRepo.DeactivateRecipientByChatID(ctx, r.TelegramChatID)
			} else {
				allSent = false
			}
		}
	}

	if allSent {
		return s.outboxRepo.MarkDelivered(ctx, item.ID)
	}
	return s.outboxRepo.MarkFailed(ctx, item.ID, "Ba'zi recipientlarga xabar yetib bormadi", true, 30*time.Second)
}

func (s *telegramService) dispatchAppointmentCancelled(ctx context.Context, item *models.NotificationOutbox) error {
	if s.clinicBotClient == nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Clinic Bot sozlanmagan")
		return nil
	}

	if item.TargetClinicID == nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "TargetClinicID yo‘q")
		return nil
	}

	recipients, err := s.telegramRepo.GetActiveClinicRecipients(ctx, *item.TargetClinicID)
	if err != nil || len(recipients) == 0 {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Ulangan faol recipientlar yo‘q")
		return nil
	}

	var data struct {
		AppointmentID string `json:"appointment_id"`
		ClinicName    string `json:"clinic_name"`
		PatientName   string `json:"patient_name"`
		DoctorName    string `json:"doctor_name"`
		Date          string `json:"date"`
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time"`
		CancelledBy   string `json:"cancelled_by"`
		Reason        string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(item.Payload), &data); err != nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Payload parslash xatoligi")
		return nil
	}

	text := fmt.Sprintf("❌ <b>QABUL BEKOR QILINDI</b>\n\n"+
		"🏥 <b>Klinika:</b> %s\n"+
		"👤 <b>Bemor:</b> %s\n"+
		"👨‍⚕️ <b>Shifokor:</b> %s\n"+
		"📅 <b>Sana:</b> %s (%s–%s)\n"+
		"⚠️ <b>Bekor qilgan:</b> %s\n"+
		"📝 <b>Sabab:</b> %s\n"+
		"🔴 <b>Holat:</b> Bekor qilingan",
		data.ClinicName, data.PatientName, data.DoctorName, data.Date, data.StartTime, data.EndTime,
		data.CancelledBy, data.Reason)

	markup := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{
					Text: "🌐 Saytga o‘tish",
					URL:  s.buildWebURL(fmt.Sprintf("/clinic/appointments/%s", data.AppointmentID)),
				},
			},
		},
	}

	allSent := true
	for _, r := range recipients {
		_, err := s.clinicBotClient.SendMessage(ctx, r.TelegramChatID, text, markup)
		if err != nil {
			if errors.Is(err, telegram.ErrBotBlockedByUser) {
				_ = s.telegramRepo.DeactivateRecipientByChatID(ctx, r.TelegramChatID)
			} else {
				allSent = false
			}
		}
	}

	if allSent {
		return s.outboxRepo.MarkDelivered(ctx, item.ID)
	}
	return s.outboxRepo.MarkFailed(ctx, item.ID, "Ba'zi recipientlarga xabar yetib bormadi", true, 30*time.Second)
}

func (s *telegramService) dispatchClinicSubmitted(ctx context.Context, item *models.NotificationOutbox) error {
	if s.adminBotClient == nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Admin Bot sozlanmagan")
		return nil
	}

	if len(s.cfg.TelegramAdminAllowedIDs) == 0 {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "TELEGRAM_ADMIN_ALLOWED_IDS belgilanmagan")
		return nil
	}

	var data struct {
		ClinicID    string `json:"clinic_id"`
		ClinicName  string `json:"clinic_name"`
		OwnerName   string `json:"owner_name"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		Region      string `json:"region"`
		District    string `json:"district"`
		SubmittedAt string `json:"submitted_at"`
	}
	if err := json.Unmarshal([]byte(item.Payload), &data); err != nil {
		_ = s.outboxRepo.MarkSkipped(ctx, item.ID, "Payload parslash xatoligi")
		return nil
	}

	text := fmt.Sprintf("🏥 <b>YANGI KLINIKA SO‘ROVI</b>\n\n"+
		"📌 <b>Klinika:</b> %s\n"+
		"👤 <b>Mas’ul shaxs:</b> %s\n"+
		"📞 <b>Telefon:</b> %s\n"+
		"📧 <b>Email:</b> %s\n"+
		"📍 <b>Hudud:</b> %s, %s\n"+
		"📅 <b>Yuborilgan sana:</b> %s\n"+
		"⏳ <b>Holat:</b> Tasdiqlash kutilmoqda",
		data.ClinicName, data.OwnerName, data.Phone, data.Email, data.Region, data.District, data.SubmittedAt)

	markup := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "✅ Tasdiqlash", CallbackData: fmt.Sprintf("cb:admin_approve:%s", data.ClinicID)},
				{Text: "❌ Rad etish", CallbackData: fmt.Sprintf("cb:admin_reject_prompt:%s", data.ClinicID)},
			},
			{
				{
					Text: "🌐 Saytga o‘tish",
					URL:  s.buildWebURL(fmt.Sprintf("/admin/clinics?clinicId=%s", data.ClinicID)),
				},
			},
		},
	}

	allSent := true
	for _, adminChatID := range s.cfg.TelegramAdminAllowedIDs {
		_, err := s.adminBotClient.SendMessage(ctx, adminChatID, text, markup)
		if err != nil {
			allSent = false
			log.Printf("Admin %d ga xabar yuborishda xatolik: %v\n", adminChatID, err)
		}
	}

	if allSent {
		return s.outboxRepo.MarkDelivered(ctx, item.ID)
	}
	return s.outboxRepo.MarkFailed(ctx, item.ID, "Superadminga xabar yetkazishda xatolik", true, 30*time.Second)
}

// ----------------------------------------------------------------------------
// 5. Long Polling Background Listeners
// ----------------------------------------------------------------------------

func (s *telegramService) StartPolling(ctx context.Context) {
	// 1. Clinic Bot initialization & health check
	if s.clinicBotClient != nil {
		me, err := s.clinicBotClient.GetMe(ctx)
		if err != nil {
			if errors.Is(err, telegram.ErrUnauthorized) {
				log.Println("❌ [Telegram Clinic Bot] 401 Unauthorized: TELEGRAM_CLINIC_BOT_TOKEN noto'g'ri!")
			} else {
				log.Printf("⚠️ [Telegram Clinic Bot] getMe tekshiruvida ogohlantirish: %v\n", err)
			}
		} else {
			log.Printf("✅ [Telegram Clinic Bot] Ulandi: @%s (%s, ID: %d)\n", me.Username, me.FirstName, me.ID)
		}
	}

	// 2. Admin Bot initialization & health check
	if s.adminBotClient != nil && s.cfg.TelegramClinicBotToken != s.cfg.TelegramAdminBotToken {
		me, err := s.adminBotClient.GetMe(ctx)
		if err != nil {
			if errors.Is(err, telegram.ErrUnauthorized) {
				log.Println("❌ [Telegram Admin Bot] 401 Unauthorized: TELEGRAM_ADMIN_BOT_TOKEN noto'g'ri!")
			} else {
				log.Printf("⚠️ [Telegram Admin Bot] getMe tekshiruvida ogohlantirish: %v\n", err)
			}
		} else {
			log.Printf("✅ [Telegram Admin Bot] Ulandi: @%s (%s, ID: %d)\n", me.Username, me.FirstName, me.ID)
		}
	}

	// 3. Prevent duplicate polling instance if identical tokens are configured (409 Conflict prevention)
	if s.cfg.TelegramClinicBotToken != "" && s.cfg.TelegramClinicBotToken == s.cfg.TelegramAdminBotToken {
		log.Println("ℹ️ [Telegram] Clinic Bot va Admin Bot uchun bir xil token sozlangan. 409 Conflict oldini olish uchun yagona birlashtirilgan polling ishga tushiriladi.")
		go s.pollUnifiedBot(ctx)
		return
	}

	if s.clinicBotClient != nil {
		go s.pollClinicBot(ctx)
	}
	if s.adminBotClient != nil {
		go s.pollAdminBot(ctx)
	}
}

func (s *telegramService) pollClinicBot(ctx context.Context) {
	log.Println("Telegram Clinic Bot polling boshlandi...")
	var offset int64 = 0

	for {
		select {
		case <-ctx.Done():
			log.Println("Telegram Clinic Bot polling to‘xtatildi.")
			return
		default:
			updates, err := s.clinicBotClient.GetUpdates(ctx, offset, 50, 25)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if errors.Is(err, telegram.ErrConflict) {
					log.Println("⚠️ [Telegram Clinic Bot] 409 Conflict: Boshqa polling instansiyasi ishlab turibdi! Iltimos, boshqa terminal yoki jarayonlarni tekshiring. 5 soniya kutilmoqda...")
					time.Sleep(5 * time.Second)
					continue
				}
				if errors.Is(err, telegram.ErrUnauthorized) {
					log.Println("❌ [Telegram Clinic Bot] 401 Unauthorized: TELEGRAM_CLINIC_BOT_TOKEN bekor qilingan yoki noto'g'ri! Polling to'xtatilmoqda.")
					return
				}
				log.Printf("⚠️ [Telegram Clinic Bot] getUpdates xatosi: %v (3 soniyadan so'ng qayta uriniladi)\n", err)
				time.Sleep(3 * time.Second)
				continue
			}

			for _, u := range updates {
				if u.UpdateID >= offset {
					offset = u.UpdateID + 1
				}
				if err := s.ProcessClinicBotUpdate(ctx, &u); err != nil {
					log.Printf("⚠️ [Telegram Clinic Bot] Update %d ni qayta ishlashda xatolik: %v\n", u.UpdateID, err)
				}
			}
		}
	}
}

func (s *telegramService) pollAdminBot(ctx context.Context) {
	log.Println("Telegram Admin Bot polling boshlandi...")
	var offset int64 = 0

	for {
		select {
		case <-ctx.Done():
			log.Println("Telegram Admin Bot polling to‘xtatildi.")
			return
		default:
			updates, err := s.adminBotClient.GetUpdates(ctx, offset, 50, 25)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if errors.Is(err, telegram.ErrConflict) {
					log.Println("⚠️ [Telegram Admin Bot] 409 Conflict: Boshqa polling instansiyasi ishlab turibdi! Iltimos, boshqa terminal yoki jarayonlarni tekshiring. 5 soniya kutilmoqda...")
					time.Sleep(5 * time.Second)
					continue
				}
				if errors.Is(err, telegram.ErrUnauthorized) {
					log.Println("❌ [Telegram Admin Bot] 401 Unauthorized: TELEGRAM_ADMIN_BOT_TOKEN bekor qilingan yoki noto'g'ri! Polling to'xtatilmoqda.")
					return
				}
				log.Printf("⚠️ [Telegram Admin Bot] getUpdates xatosi: %v (3 soniyadan so'ng qayta uriniladi)\n", err)
				time.Sleep(3 * time.Second)
				continue
			}

			for _, u := range updates {
				if u.UpdateID >= offset {
					offset = u.UpdateID + 1
				}
				if err := s.ProcessAdminBotUpdate(ctx, &u); err != nil {
					log.Printf("⚠️ [Telegram Admin Bot] Update %d ni qayta ishlashda xatolik: %v\n", u.UpdateID, err)
				}
			}
		}
	}
}

func (s *telegramService) pollUnifiedBot(ctx context.Context) {
	log.Println("Telegram Unified (Clinic + Admin) Bot polling boshlandi...")
	var offset int64 = 0

	for {
		select {
		case <-ctx.Done():
			log.Println("Telegram Unified Bot polling to‘xtatildi.")
			return
		default:
			updates, err := s.clinicBotClient.GetUpdates(ctx, offset, 50, 25)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if errors.Is(err, telegram.ErrConflict) {
					log.Println("⚠️ [Telegram Unified Bot] 409 Conflict: Boshqa polling instansiyasi ishlab turibdi! 5 soniya kutilmoqda...")
					time.Sleep(5 * time.Second)
					continue
				}
				if errors.Is(err, telegram.ErrUnauthorized) {
					log.Println("❌ [Telegram Unified Bot] 401 Unauthorized: Bot tokeni noto'g'ri! Polling to'xtatilmoqda.")
					return
				}
				log.Printf("⚠️ [Telegram Unified Bot] getUpdates xatosi: %v (3 soniyadan so'ng qayta uriniladi)\n", err)
				time.Sleep(3 * time.Second)
				continue
			}

			for _, u := range updates {
				if u.UpdateID >= offset {
					offset = u.UpdateID + 1
				}
				// Route update to admin or clinic
				if u.CallbackQuery != nil && strings.HasPrefix(u.CallbackQuery.Data, "cb:admin_") {
					_ = s.ProcessAdminBotUpdate(ctx, &u)
				} else if u.Message != nil && (strings.HasPrefix(strings.TrimSpace(u.Message.Text), "/pending") ||
					strings.HasPrefix(strings.TrimSpace(u.Message.Text), "/approved") ||
					s.hasAdminSession(ctx, u.Message.From)) {
					_ = s.ProcessAdminBotUpdate(ctx, &u)
				} else {
					_ = s.ProcessClinicBotUpdate(ctx, &u)
				}
			}
		}
	}
}

func (s *telegramService) hasAdminSession(ctx context.Context, user *telegram.User) bool {
	if user == nil {
		return false
	}
	session, err := s.telegramRepo.GetAdminSession(ctx, user.ID)
	return err == nil && session != nil && session.AwaitingAction != ""
}

// Helpers
func (s *telegramService) buildWebURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.cfg.AppPublicURL), "/")
	if baseURL == "" || strings.Contains(baseURL, "localhost") || strings.Contains(baseURL, "127.0.0.1") || strings.Contains(baseURL, "::1") {
		baseURL = "https://qabulgo.uz"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return baseURL + path
}

func (s *telegramService) getTashkentDateOffset(dayOffset int) string {
	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		loc = time.FixedZone("UTC+5", 5*3600)
	}
	target := time.Now().In(loc).AddDate(0, 0, dayOffset)
	return target.Format("2006-01-02")
}
