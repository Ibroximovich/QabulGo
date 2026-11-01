package routes

import (
	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/handlers"
	"qabulgo-backend/internal/middleware"
	"qabulgo-backend/internal/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

type RouterConfig struct {
	Config           *config.Config
	HealthHandler    *handlers.HealthHandler
	AuthHandler      *handlers.AuthHandler
	PatientHandler   *handlers.PatientHandler
	ClinicHandler    *handlers.ClinicHandler
	DoctorHandler    *handlers.DoctorHandler
	AdminHandler     *handlers.AdminHandler
	SpecialtyHandler    *handlers.SpecialtyHandler
	UploadHandler       *handlers.UploadHandler
	PublicSearchHandler *handlers.PublicSearchHandler
}

// SetupRouter initializes and configures the Gin engine with all routes and middlewares
func SetupRouter(rc *RouterConfig) *gin.Engine {
	if rc.Config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Rate limiter
	rateLimiter := middleware.NewIPRateLimiter(
		rate.Limit(rc.Config.RateLimitRPS),
		rc.Config.RateLimitBurst,
	)

	// Global Middlewares
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(rc.Config))
	r.Use(middleware.RateLimit(rateLimiter))

	// Static uploads directory
	r.Static("/uploads", "./uploads")

	// Health Check
	r.GET("/health", rc.HealthHandler.HealthCheck)

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 1. Authentication Routes
		auth := v1.Group("/auth")
		{
			// Patient Auth
			auth.POST("/patient/register", rc.AuthHandler.Register)
			auth.POST("/patient/login", rc.AuthHandler.Login)

			// Clinic Auth
			auth.POST("/clinic/register", rc.AuthHandler.RegisterClinic)
			auth.POST("/clinic/login", rc.AuthHandler.LoginClinic)

			// Superadmin Auth
			auth.POST("/admin/login", rc.AuthHandler.LoginSuperAdmin)

			// Shared Auth utilities
			auth.POST("/verify-email", rc.AuthHandler.VerifyEmail)
			auth.POST("/resend-verification", rc.AuthHandler.ResendVerification)
			auth.POST("/refresh", rc.AuthHandler.Refresh)
			auth.POST("/logout", rc.AuthHandler.Logout)
			auth.POST("/forgot-password", rc.AuthHandler.ForgotPassword)
			auth.POST("/reset-password", rc.AuthHandler.ResetPassword)
		}

		// 2. Public Specialties Route
		if rc.SpecialtyHandler != nil {
			v1.GET("/specialties", rc.SpecialtyHandler.ListSpecialties)
		}

		// 3. Protected Patient Routes
		patient := v1.Group("/patient")
		patient.Use(middleware.Auth(rc.Config))
		patient.Use(middleware.RequireRole(models.RolePatient))
		{
			patient.GET("/me", rc.PatientHandler.GetProfile)
			patient.PATCH("/me", rc.PatientHandler.UpdateProfile)
		}

		// 4. File Upload (Authenticated Users: ClinicOwner, ClinicAdmin, SuperAdmin)
		if rc.UploadHandler != nil {
			upload := v1.Group("/upload")
			upload.Use(middleware.Auth(rc.Config))
			upload.Use(middleware.RequireRole(models.RoleClinicOwner, models.RoleClinicAdmin, models.RoleSuperAdmin))
			{
				upload.POST("", rc.UploadHandler.UploadImage)
			}
		}

		// 5. Clinic Profile & Management Routes
		if rc.ClinicHandler != nil {
			clinics := v1.Group("/clinics")
			{
				// Public / General View (status filtered inside handler)
				clinics.GET("/:id", rc.ClinicHandler.GetClinicByID)

				// Authenticated Clinic Owner & Admin Routes
				protectedClinics := clinics.Group("")
				protectedClinics.Use(middleware.Auth(rc.Config))
				protectedClinics.Use(middleware.RequireRole(models.RoleClinicOwner, models.RoleClinicAdmin))
				{
					protectedClinics.POST("", rc.ClinicHandler.CreateClinic)
					protectedClinics.GET("/mine", rc.ClinicHandler.GetMyClinic)
					protectedClinics.PATCH("/:id", rc.ClinicHandler.UpdateClinic)
					protectedClinics.POST("/:id/submit", rc.ClinicHandler.SubmitClinic)
				}
			}
		}

		// 6. Doctor Management Routes
		if rc.DoctorHandler != nil {
			clinicsDoctors := v1.Group("/clinics/:id/doctors")
			{
				// Read doctors (Public or Authenticated)
				clinicsDoctors.GET("", rc.DoctorHandler.ListDoctors)
				clinicsDoctors.GET("/:doctorId", rc.DoctorHandler.GetDoctor)

				// Manage doctors (Clinic Owner & Admin only)
				protectedDoctors := clinicsDoctors.Group("")
				protectedDoctors.Use(middleware.Auth(rc.Config))
				protectedDoctors.Use(middleware.RequireRole(models.RoleClinicOwner, models.RoleClinicAdmin))
				{
					protectedDoctors.POST("", rc.DoctorHandler.AddDoctor)
					protectedDoctors.PATCH("/:doctorId", rc.DoctorHandler.UpdateDoctor)
					protectedDoctors.DELETE("/:doctorId", rc.DoctorHandler.DeactivateDoctor)
				}
			}
		}

		// 7. Superadmin Routes
		if rc.AdminHandler != nil {
			admin := v1.Group("/admin")
			admin.Use(middleware.Auth(rc.Config))
			admin.Use(middleware.RequireRole(models.RoleSuperAdmin))
			{
				admin.GET("/clinics", rc.AdminHandler.ListClinics)
				admin.GET("/clinics/:id", rc.AdminHandler.GetClinicDetails)
				admin.POST("/clinics/:id/approve", rc.AdminHandler.ApproveClinic)
				admin.POST("/clinics/:id/reject", rc.AdminHandler.RejectClinic)
				admin.POST("/clinics/:id/suspend", rc.AdminHandler.SuspendClinic)
			}
		}

		// 8. Public Search & Geolocation Routes (Public / Unauthenticated)
		if rc.PublicSearchHandler != nil {
			public := v1.Group("/public")
			{
				// Clinics
				// MUHIM: /clinics/nearby static route parametrli /clinics/:id dan oldin ro'yxatdan o'tishi shart!
				public.GET("/clinics/nearby", rc.PublicSearchHandler.SearchNearbyClinics)
				public.GET("/clinics", rc.PublicSearchHandler.SearchClinics)
				public.GET("/clinics/:id", rc.PublicSearchHandler.GetClinicDetails)
				public.GET("/clinics/:id/doctors", rc.PublicSearchHandler.GetClinicDoctors)

				// Doctors
				public.GET("/doctors", rc.PublicSearchHandler.SearchDoctors)
				public.GET("/doctors/:id", rc.PublicSearchHandler.GetDoctorDetails)

				// Regions catalog
				public.GET("/regions", rc.PublicSearchHandler.GetRegionsCatalog)
			}
		}
	}

	return r
}
