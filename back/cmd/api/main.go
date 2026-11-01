package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/database"
	"qabulgo-backend/internal/handlers"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/routes"
	"qabulgo-backend/internal/services"

	_ "qabulgo-backend/docs" // Swagger docs registration
)

// @title QabulGo Backend API
// @version 1.0
// @description QabulGo — O‘zbekistondagi xususiy klinikalar, shifokorlar va bemorlarni bog‘lovchi platforma (1-bosqich: Bemorlar uchun Authentication).
// @contact.name QabulGo Support
// @contact.email support@qabulgo.uz
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	log.Println("Starting QabulGo Backend Service...")

	// 1. Konfiguratsiyani yuklash
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Konfiguratsiyani yuklashda xatolik: %v", err)
	}

	// 2. Ma’lumotlar bazasiga ulanish
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Ma’lumotlar bazasiga ulanishda xatolik yuz berdi: %v\nIltimos, PostgreSQL ishlab turganligini va konfiguratsiyani tekshiring.", err)
	}

	// Migratsiyalarni amalga oshirish
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("sql.DB instansiyasini olishda xatolik: %v", err)
	}
	if err := database.RunMigrationsUp(sqlDB, "migrations"); err != nil {
		log.Fatalf("Ma’lumotlar bazasi migratsiyalarini bajarishda xatolik: %v", err)
	}

	// 3. Repositories initsializatsiyasi
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewRefreshSessionRepository(db)
	tokenRepo := repositories.NewVerificationTokenRepository(db)
	clinicRepo := repositories.NewClinicRepository(db)
	doctorRepo := repositories.NewDoctorRepository(db)
	specialtyRepo := repositories.NewSpecialtyRepository(db)
	auditRepo := repositories.NewAuditRepository(db)
	publicSearchRepo := repositories.NewPublicSearchRepository(db)
	scheduleRepo := repositories.NewScheduleRepository(db)

	// 4. Services initsializatsiyasi
	emailService := services.NewEmailService(cfg)
	tokenService := services.NewTokenService(cfg, sessionRepo, userRepo)
	authService := services.NewAuthService(cfg, userRepo, tokenRepo, tokenService, emailService)
	patientService := services.NewPatientService(userRepo)
	clinicService := services.NewClinicService(clinicRepo, doctorRepo, userRepo, auditRepo)
	doctorService := services.NewDoctorService(doctorRepo, clinicRepo, specialtyRepo, auditRepo)
	adminService := services.NewAdminService(clinicRepo, doctorRepo, auditRepo)
	specialtyService := services.NewSpecialtyService(specialtyRepo)
	storageService := services.NewLocalStorageService("uploads", "/uploads")
	publicSearchService := services.NewPublicSearchService(publicSearchRepo, doctorRepo)
	scheduleService := services.NewScheduleService(scheduleRepo, clinicRepo, auditRepo)

	// Mutaxassisliklar boshlang'ich seed
	_ = specialtyService.SeedIfEmpty(context.Background())

	// 5. Handlers initsializatsiyasi
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg, authService, tokenService)
	patientHandler := handlers.NewPatientHandler(patientService)
	clinicHandler := handlers.NewClinicHandler(clinicService)
	doctorHandler := handlers.NewDoctorHandler(doctorService)
	adminHandler := handlers.NewAdminHandler(adminService)
	specialtyHandler := handlers.NewSpecialtyHandler(specialtyService)
	uploadHandler := handlers.NewUploadHandler(storageService)
	publicSearchHandler := handlers.NewPublicSearchHandler(publicSearchService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)

	// 6. Router initsializatsiyasi
	router := routes.SetupRouter(&routes.RouterConfig{
		Config:              cfg,
		HealthHandler:       healthHandler,
		AuthHandler:         authHandler,
		PatientHandler:      patientHandler,
		ClinicHandler:       clinicHandler,
		DoctorHandler:       doctorHandler,
		AdminHandler:        adminHandler,
		SpecialtyHandler:    specialtyHandler,
		UploadHandler:       uploadHandler,
		PublicSearchHandler: publicSearchHandler,
		ScheduleHandler:     scheduleHandler,
	})

	// 7. HTTP Server sozlamalari
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Graceful Shutdown kanali
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("QabulGo Backend server http://localhost:%s da ishga tushdi\n", cfg.Port)
		log.Printf("Swagger API hujjatlari: http://localhost:%s/swagger/index.html\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Serverni ishga tushirishda xatolik: %v", err)
		}
	}()

	// Signalni kutish
	<-shutdownChan
	log.Println("Serverni to‘xtatish so‘rovi qabul qilindi. Graceful shutdown boshlanmoqda...")

	// Kontekst bilan serverni xavfsiz to‘xtatish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Serverni to‘xtatishda xatolik yuz berdi: %v", err)
	}

	log.Println("QabulGo Backend server muvaffaqiyatli to‘xtatildi.")
}
