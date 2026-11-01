package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/database"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

func main() {
	emailFlag := flag.String("email", "", "Superadmin email manzili (majburiy)")
	passwordFlag := flag.String("password", "", "Superadmin paroli (kamida 8 ta belgi)")
	firstNameFlag := flag.String("firstName", "Super", "Ism")
	lastNameFlag := flag.String("lastName", "Admin", "Familiya")
	phoneFlag := flag.String("phone", "+998900000000", "Telefon raqami")
	flag.Parse()

	if *emailFlag == "" || *passwordFlag == "" {
		fmt.Println("Foydalanish: go run cmd/admin/main.go -email admin@qabulgo.uz -password \"SuperAdminPass123!\"")
		os.Exit(1)
	}

	if len(*passwordFlag) < 8 {
		log.Fatalf("Parol uzunligi kamida 8 ta belgidan iborat bo‘lishi shart!")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Konfiguratsiyani yuklashda xatolik: %v", err)
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Ma’lumotlar bazasiga ulanishda xatolik: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	ctx := context.Background()

	normalizedEmail := utils.NormalizeEmail(*emailFlag)

	// Agar foydalanuvchi mavjud bo'lsa, rolini super_admin ga yangilaymiz
	existingUser, err := userRepo.GetByEmail(ctx, normalizedEmail)
	if err == nil && existingUser != nil {
		log.Printf("Foydalanuvchi %s bazada topildi. Paroli va roli super_admin ga yangilanmoqda...", normalizedEmail)
		hash, err := utils.HashPassword(*passwordFlag)
		if err != nil {
			log.Fatalf("Parolni xeshlashda xatolik: %v", err)
		}

		existingUser.PasswordHash = hash
		existingUser.Role = models.RoleSuperAdmin
		existingUser.Status = models.StatusActive
		now := time.Now().UTC()
		existingUser.EmailVerifiedAt = &now

		if err := userRepo.Update(ctx, existingUser); err != nil {
			log.Fatalf("Foydalanuvchini yangilashda xatolik: %v", err)
		}
		log.Printf("Muvaffaqiyatli! %s foydalanuvchisi super_admin roliga o‘tkazildi.", normalizedEmail)
		return
	}

	// Yangi superadmin yaratamiz
	hash, err := utils.HashPassword(*passwordFlag)
	if err != nil {
		log.Fatalf("Parolni xeshlashda xatolik: %v", err)
	}

	now := time.Now().UTC()
	superAdmin := &models.User{
		ID:              uuid.New(),
		FirstName:       utils.SanitizeString(*firstNameFlag),
		LastName:        utils.SanitizeString(*lastNameFlag),
		Email:           normalizedEmail,
		Phone:           strings.TrimSpace(*phoneFlag),
		PasswordHash:    hash,
		Role:            models.RoleSuperAdmin,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
		TermsAcceptedAt: now,
		TermsVersion:    "v1.0",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := userRepo.Create(ctx, superAdmin); err != nil {
		log.Fatalf("Superadmin yaratishda xatolik: %v", err)
	}

	log.Printf("Muvaffaqiyatli! Yangi superadmin yaratildi: %s (ID: %s)", normalizedEmail, superAdmin.ID.String())
}

