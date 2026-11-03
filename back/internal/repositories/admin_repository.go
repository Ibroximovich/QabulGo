package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPatientNotFound = errors.New("bemor topilmadi")
)

type AdminRepository interface {
	GetDashboardSummary(ctx context.Context) (*dto.AdminDashboardSummaryDTO, error)
	GetDashboardAnalytics(ctx context.Context, filter *dto.AdminDashboardAnalyticsFilter) (*dto.AdminDashboardAnalyticsDTO, error)
	GetClinicDoctors(ctx context.Context, clinicID uuid.UUID) ([]dto.AdminClinicDoctorItemDTO, error)
	GetClinicAppointments(ctx context.Context, clinicID uuid.UUID, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error)
	GetClinicAnalytics(ctx context.Context, clinicID uuid.UUID) (*dto.AdminClinicAnalyticsDTO, error)
	ListPatients(ctx context.Context, filter *dto.AdminPatientFilter) ([]dto.AdminPatientListItemDTO, int64, error)
	GetPatientDetails(ctx context.Context, patientID uuid.UUID) (*dto.AdminPatientDetailDTO, error)
	ListDoctors(ctx context.Context, filter *dto.AdminDoctorFilter) ([]dto.AdminDoctorListItemDTO, int64, error)
	GetDoctorDetails(ctx context.Context, doctorID uuid.UUID) (*dto.AdminDoctorDetailDTO, error)
	ListAppointments(ctx context.Context, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error)
	GetAppointmentDetails(ctx context.Context, appointmentID uuid.UUID) (*dto.AdminAppointmentListItemDTO, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetDashboardSummary(ctx context.Context) (*dto.AdminDashboardSummaryDTO, error) {
	summary := &dto.AdminDashboardSummaryDTO{}

	// 1. Clinics stats
	_ = r.db.WithContext(ctx).Model(&models.Clinic{}).Count(&summary.TotalClinics)
	_ = r.db.WithContext(ctx).Model(&models.Clinic{}).Where("status = ?", models.ClinicStatusApproved).Count(&summary.ApprovedClinics)
	_ = r.db.WithContext(ctx).Model(&models.Clinic{}).Where("status = ?", models.ClinicStatusPending).Count(&summary.PendingClinics)
	_ = r.db.WithContext(ctx).Model(&models.Clinic{}).Where("status = ?", models.ClinicStatusRejected).Count(&summary.RejectedClinics)
	_ = r.db.WithContext(ctx).Model(&models.Clinic{}).Where("status = ?", models.ClinicStatusSuspended).Count(&summary.SuspendedClinics)

	// 2. Total Patients
	_ = r.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", models.RolePatient).Count(&summary.TotalPatients)

	// 3. Total Unique Doctors
	_ = r.db.WithContext(ctx).Model(&models.Doctor{}).Count(&summary.TotalDoctors)

	// 4. Appointments Total & Statuses
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Count(&summary.TotalAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("status = ?", models.AppointmentStatusConfirmed).Count(&summary.ConfirmedAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("status = ?", models.AppointmentStatusCancelled).Count(&summary.CancelledAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("status = ?", models.AppointmentStatusCompleted).Count(&summary.CompletedAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("status = ?", models.AppointmentStatusNoShow).Count(&summary.NoShowAppointments)

	// 5. Time-scoped appointments
	todayStr := time.Now().UTC().Format("2006-01-02")
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("appointment_date = ?", todayStr).Count(&summary.TodayAppointments)

	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())).Format("2006-01-02")
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("appointment_date >= ?", weekStart).Count(&summary.ThisWeekAppointments)

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("appointment_date >= ?", monthStart).Count(&summary.ThisMonthAppointments)

	// 6. Estimated total revenue of all bookings
	type RevResult struct {
		Total float64
	}
	var revResult RevResult
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).
		Select("COALESCE(SUM(consultation_price), 0) as total").
		Scan(&revResult)
	summary.EstimatedTotalRevenue = revResult.Total

	return summary, nil
}

func (r *adminRepository) GetDashboardAnalytics(ctx context.Context, filter *dto.AdminDashboardAnalyticsFilter) (*dto.AdminDashboardAnalyticsDTO, error) {
	analytics := &dto.AdminDashboardAnalyticsDTO{
		DailyBookings:      make([]dto.DailyMetricDTO, 0),
		WeeklyBookings:     make([]dto.DailyMetricDTO, 0),
		MonthlyBookings:    make([]dto.DailyMetricDTO, 0),
		ClinicGrowth:       make([]dto.DailyMetricDTO, 0),
		PatientGrowth:      make([]dto.DailyMetricDTO, 0),
		StatusDistribution: make([]dto.StatusCountDTO, 0),
		TopClinics:         make([]dto.TopClinicMetricDTO, 0),
	}

	startDate := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	switch filter.Range {
	case "today":
		startDate = time.Now().UTC().Format("2006-01-02")
	case "7d":
		startDate = time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02")
	case "30d":
		startDate = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	case "90d":
		startDate = time.Now().UTC().AddDate(0, 0, -90).Format("2006-01-02")
	case "custom":
		if filter.StartDate != "" {
			startDate = filter.StartDate
		}
	}

	// 1. Daily Bookings in date range
	_ = r.db.WithContext(ctx).Table("appointments").
		Select("appointment_date as date, COUNT(*) as count, COALESCE(SUM(consultation_price), 0) as amount").
		Where("appointment_date >= ?", startDate).
		Group("appointment_date").
		Order("appointment_date ASC").
		Scan(&analytics.DailyBookings)

	// 2. Status distribution
	_ = r.db.WithContext(ctx).Table("appointments").
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&analytics.StatusDistribution)

	// 3. Top clinics by bookings
	type topClinicRow struct {
		ClinicID      uuid.UUID
		ClinicName    string
		BookingsCount int64
		TotalRevenue  float64
	}
	var topRows []topClinicRow
	_ = r.db.WithContext(ctx).Table("appointments a").
		Joins("JOIN clinics c ON c.id = a.clinic_id").
		Select("c.id as clinic_id, c.name as clinic_name, COUNT(a.id) as bookings_count, COALESCE(SUM(a.consultation_price), 0) as total_revenue").
		Group("c.id, c.name").
		Order("bookings_count DESC").
		Limit(5).
		Scan(&topRows)

	for _, tr := range topRows {
		analytics.TopClinics = append(analytics.TopClinics, dto.TopClinicMetricDTO{
			ClinicID:      tr.ClinicID,
			ClinicName:    tr.ClinicName,
			BookingsCount: tr.BookingsCount,
			TotalRevenue:  tr.TotalRevenue,
		})
	}

	// 4. Clinic growth by creation date
	_ = r.db.WithContext(ctx).Table("clinics").
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&analytics.ClinicGrowth)

	// 5. Patient growth by creation date
	_ = r.db.WithContext(ctx).Table("users").
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("role = ? AND created_at >= ?", models.RolePatient, startDate).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&analytics.PatientGrowth)

	return analytics, nil
}

func (r *adminRepository) ListPatients(ctx context.Context, filter *dto.AdminPatientFilter) ([]dto.AdminPatientListItemDTO, int64, error) {
	filter.Normalize()

	query := r.db.WithContext(ctx).Table("users u").
		Where("u.role = ?", models.RolePatient)

	if filter.Search != "" {
		s := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(u.first_name) LIKE ? OR LOWER(u.last_name) LIKE ? OR LOWER(u.email) LIKE ? OR u.phone LIKE ?", s, s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	type PatientRow struct {
		ID                  uuid.UUID
		FirstName           string
		LastName            string
		Email               string
		Phone               string
		Status              string
		CreatedAt           time.Time
		AppointmentsCount   int64
		LastAppointmentDate *string
	}

	var rows []PatientRow
	err := query.Select(`
		u.id, u.first_name, u.last_name, u.email, u.phone, u.status, u.created_at,
		(SELECT COUNT(*) FROM appointments a WHERE a.patient_id = u.id) as appointments_count,
		(SELECT MAX(a.appointment_date) FROM appointments a WHERE a.patient_id = u.id) as last_appointment_date
	`).
		Order("u.created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	items := make([]dto.AdminPatientListItemDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.AdminPatientListItemDTO{
			ID:                  row.ID,
			FirstName:           row.FirstName,
			LastName:            row.LastName,
			Email:               row.Email,
			Phone:               row.Phone,
			Status:              row.Status,
			CreatedAt:           row.CreatedAt,
			AppointmentsCount:   row.AppointmentsCount,
			LastAppointmentDate: row.LastAppointmentDate,
		})
	}

	return items, total, nil
}

func (r *adminRepository) GetPatientDetails(ctx context.Context, patientID uuid.UUID) (*dto.AdminPatientDetailDTO, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ? AND role = ?", patientID, models.RolePatient).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	detail := &dto.AdminPatientDetailDTO{
		ID:                 user.ID,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Email:              user.Email,
		Phone:              user.Phone,
		Status:             string(user.Status),
		CreatedAt:          user.CreatedAt,
		RecentAppointments: make([]dto.AdminAppointmentListItemDTO, 0),
	}

	// Counts
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("patient_id = ?", patientID).Count(&detail.TotalAppointments)
	todayStr := time.Now().UTC().Format("2006-01-02")
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("patient_id = ? AND status = ? AND appointment_date >= ?", patientID, models.AppointmentStatusConfirmed, todayStr).Count(&detail.UpcomingAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("patient_id = ? AND status = ?", patientID, models.AppointmentStatusCompleted).Count(&detail.CompletedAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("patient_id = ? AND status = ?", patientID, models.AppointmentStatusCancelled).Count(&detail.CancelledAppointments)

	// Recent appointments
	recentFilter := &dto.AdminAppointmentFilter{Page: 1, Limit: 10}
	recentAppointments, _, err := r.getAppointmentsQuery(ctx, r.db.Where("a.patient_id = ?", patientID), recentFilter)
	if err == nil {
		detail.RecentAppointments = recentAppointments
	}

	return detail, nil
}

func (r *adminRepository) ListDoctors(ctx context.Context, filter *dto.AdminDoctorFilter) ([]dto.AdminDoctorListItemDTO, int64, error) {
	filter.Normalize()

	query := r.db.WithContext(ctx).Table("doctors d")

	if filter.Search != "" {
		s := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(d.first_name) LIKE ? OR LOWER(d.last_name) LIKE ?", s, s)
	}

	if filter.SpecialtyID != "" {
		query = query.Where("EXISTS (SELECT 1 FROM clinic_doctors cd WHERE cd.doctor_id = d.id AND cd.specialty_id = ? AND cd.deleted_at IS NULL)", filter.SpecialtyID)
	}

	if filter.ClinicID != "" {
		query = query.Where("EXISTS (SELECT 1 FROM clinic_doctors cd WHERE cd.doctor_id = d.id AND cd.clinic_id = ? AND cd.deleted_at IS NULL)", filter.ClinicID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	var doctors []models.Doctor
	if err := query.Order("d.created_at DESC").Limit(filter.Limit).Offset(offset).Find(&doctors).Error; err != nil {
		return nil, 0, err
	}

	items := make([]dto.AdminDoctorListItemDTO, 0, len(doctors))
	for _, doc := range doctors {
		item := dto.AdminDoctorListItemDTO{
			ID:              doc.ID,
			FirstName:       doc.FirstName,
			LastName:        doc.LastName,
			Bio:             doc.Bio,
			ExperienceYears: doc.ExperienceYears,
			PhotoURL:        doc.PhotoURL,
			Specialties:     make([]string, 0),
			Clinics:         make([]string, 0),
			CreatedAt:       doc.CreatedAt,
		}

		// Count total appointments for this doctor
		_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("doctor_id = ?", doc.ID).Count(&item.TotalAppointments)

		// Get clinics and specialties
		type DocRelRow struct {
			ClinicName    string
			SpecialtyName string
		}
		var relRows []DocRelRow
		_ = r.db.WithContext(ctx).Table("clinic_doctors cd").
			Joins("JOIN clinics c ON c.id = cd.clinic_id").
			Joins("JOIN specialties s ON s.id = cd.specialty_id").
			Where("cd.doctor_id = ? AND cd.deleted_at IS NULL", doc.ID).
			Select("c.name as clinic_name, s.name as specialty_name").
			Scan(&relRows)

		seenClinics := make(map[string]bool)
		seenSpecialties := make(map[string]bool)
		for _, row := range relRows {
			if !seenClinics[row.ClinicName] {
				seenClinics[row.ClinicName] = true
				item.Clinics = append(item.Clinics, row.ClinicName)
			}
			if !seenSpecialties[row.SpecialtyName] {
				seenSpecialties[row.SpecialtyName] = true
				item.Specialties = append(item.Specialties, row.SpecialtyName)
			}
		}
		item.ClinicsCount = len(item.Clinics)

		items = append(items, item)
	}

	return items, total, nil
}

func (r *adminRepository) GetDoctorDetails(ctx context.Context, doctorID uuid.UUID) (*dto.AdminDoctorDetailDTO, error) {
	var doc models.Doctor
	if err := r.db.WithContext(ctx).Where("id = ?", doctorID).First(&doc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	detail := &dto.AdminDoctorDetailDTO{
		ID:                    doc.ID,
		FirstName:             doc.FirstName,
		LastName:              doc.LastName,
		Bio:                   doc.Bio,
		ExperienceYears:       doc.ExperienceYears,
		PhotoURL:              doc.PhotoURL,
		CreatedAt:             doc.CreatedAt,
		Clinics:               make([]dto.AdminDoctorClinicInfoDTO, 0),
		RecentAppointments:    make([]dto.AdminAppointmentListItemDTO, 0),
	}

	// Attached clinics
	type ClinDocRow struct {
		ClinicID          uuid.UUID
		ClinicName        string
		ClinicDoctorID    uuid.UUID
		SpecialtyName     string
		ConsultationPrice float64
		IsActive          bool
	}
	var clinRows []ClinDocRow
	_ = r.db.WithContext(ctx).Table("clinic_doctors cd").
		Joins("JOIN clinics c ON c.id = cd.clinic_id").
		Joins("JOIN specialties s ON s.id = cd.specialty_id").
		Where("cd.doctor_id = ? AND cd.deleted_at IS NULL", doctorID).
		Select("c.id as clinic_id, c.name as clinic_name, cd.id as clinic_doctor_id, s.name as specialty_name, cd.consultation_price, cd.is_active").
		Scan(&clinRows)

	for _, cr := range clinRows {
		detail.Clinics = append(detail.Clinics, dto.AdminDoctorClinicInfoDTO{
			ClinicID:          cr.ClinicID,
			ClinicName:        cr.ClinicName,
			ClinicDoctorID:    cr.ClinicDoctorID,
			SpecialtyName:     cr.SpecialtyName,
			ConsultationPrice: cr.ConsultationPrice,
			IsActive:          cr.IsActive,
		})
	}

	// Stats
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("doctor_id = ?", doctorID).Count(&detail.TotalAppointments)
	todayStr := time.Now().UTC().Format("2006-01-02")
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("doctor_id = ? AND status = ? AND appointment_date >= ?", doctorID, models.AppointmentStatusConfirmed, todayStr).Count(&detail.UpcomingAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("doctor_id = ? AND status = ?", doctorID, models.AppointmentStatusCompleted).Count(&detail.CompletedAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("doctor_id = ? AND status = ?", doctorID, models.AppointmentStatusCancelled).Count(&detail.CancelledAppointments)

	// Recent appointments
	recentFilter := &dto.AdminAppointmentFilter{Page: 1, Limit: 10}
	recentAppointments, _, err := r.getAppointmentsQuery(ctx, r.db.Where("a.doctor_id = ?", doctorID), recentFilter)
	if err == nil {
		detail.RecentAppointments = recentAppointments
	}

	return detail, nil
}

func (r *adminRepository) ListAppointments(ctx context.Context, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error) {
	return r.getAppointmentsQuery(ctx, r.db, filter)
}

func (r *adminRepository) GetAppointmentDetails(ctx context.Context, appointmentID uuid.UUID) (*dto.AdminAppointmentListItemDTO, error) {
	filter := &dto.AdminAppointmentFilter{Page: 1, Limit: 1}
	items, _, err := r.getAppointmentsQuery(ctx, r.db.Where("a.id = ?", appointmentID), filter)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrAppointmentNotFound
	}
	return &items[0], nil
}

func (r *adminRepository) GetClinicDoctors(ctx context.Context, clinicID uuid.UUID) ([]dto.AdminClinicDoctorItemDTO, error) {
	type CDocRow struct {
		ClinicDoctorID             uuid.UUID
		DoctorID                   uuid.UUID
		FirstName                  string
		LastName                   string
		SpecialtyName              string
		ConsultationPrice          float64
		AppointmentDurationMinutes int
		ExperienceYears            int
		IsActive                   bool
		PhotoURL                   string
	}
	var rows []CDocRow
	err := r.db.WithContext(ctx).Table("clinic_doctors cd").
		Joins("JOIN doctors d ON d.id = cd.doctor_id").
		Joins("JOIN specialties s ON s.id = cd.specialty_id").
		Where("cd.clinic_id = ? AND cd.deleted_at IS NULL", clinicID).
		Select("cd.id as clinic_doctor_id, d.id as doctor_id, d.first_name, d.last_name, s.name as specialty_name, cd.consultation_price, cd.appointment_duration_minutes, d.experience_years, cd.is_active, d.photo_url").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	items := make([]dto.AdminClinicDoctorItemDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.AdminClinicDoctorItemDTO{
			ClinicDoctorID:             row.ClinicDoctorID,
			DoctorID:                   row.DoctorID,
			FirstName:                  row.FirstName,
			LastName:                   row.LastName,
			SpecialtyName:              row.SpecialtyName,
			ConsultationPrice:          row.ConsultationPrice,
			AppointmentDurationMinutes: row.AppointmentDurationMinutes,
			ExperienceYears:            row.ExperienceYears,
			IsActive:                   row.IsActive,
			PhotoURL:                   row.PhotoURL,
		})
	}

	return items, nil
}

func (r *adminRepository) GetClinicAppointments(ctx context.Context, clinicID uuid.UUID, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error) {
	return r.getAppointmentsQuery(ctx, r.db.Where("a.clinic_id = ?", clinicID), filter)
}

func (r *adminRepository) GetClinicAnalytics(ctx context.Context, clinicID uuid.UUID) (*dto.AdminClinicAnalyticsDTO, error) {
	analytics := &dto.AdminClinicAnalyticsDTO{
		StatusDistribution: make([]dto.StatusCountDTO, 0),
		TopDoctors:         make([]dto.AdminClinicTopDoctorDTO, 0),
	}

	// Counts
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("clinic_id = ?", clinicID).Count(&analytics.TotalAppointments)
	todayStr := time.Now().UTC().Format("2006-01-02")
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("clinic_id = ? AND appointment_date = ?", clinicID, todayStr).Count(&analytics.TodayAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("clinic_id = ? AND status = ? AND appointment_date >= ?", clinicID, models.AppointmentStatusConfirmed, todayStr).Count(&analytics.UpcomingAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("clinic_id = ? AND status = ?", clinicID, models.AppointmentStatusCompleted).Count(&analytics.CompletedAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("clinic_id = ? AND status = ?", clinicID, models.AppointmentStatusCancelled).Count(&analytics.CancelledAppointments)
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).Where("clinic_id = ? AND status = ?", clinicID, models.AppointmentStatusNoShow).Count(&analytics.NoShowAppointments)

	// Revenue
	type RevResult struct {
		Total float64
	}
	var revResult RevResult
	_ = r.db.WithContext(ctx).Model(&models.Appointment{}).
		Where("clinic_id = ?", clinicID).
		Select("COALESCE(SUM(consultation_price), 0) as total").
		Scan(&revResult)
	analytics.EstimatedTotalRevenue = revResult.Total

	// Status distribution
	_ = r.db.WithContext(ctx).Table("appointments").
		Where("clinic_id = ?", clinicID).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&analytics.StatusDistribution)

	// Top doctors in clinic
	type TopDocRow struct {
		DoctorID      uuid.UUID
		DoctorName    string
		SpecialtyName string
		BookingsCount int64
	}
	var topRows []TopDocRow
	_ = r.db.WithContext(ctx).Table("appointments a").
		Joins("JOIN doctors d ON d.id = a.doctor_id").
		Joins("JOIN clinic_doctors cd ON cd.id = a.clinic_doctor_id").
		Joins("JOIN specialties s ON s.id = cd.specialty_id").
		Where("a.clinic_id = ?", clinicID).
		Select("d.id as doctor_id, CONCAT(d.first_name, ' ', d.last_name) as doctor_name, s.name as specialty_name, COUNT(a.id) as bookings_count").
		Group("d.id, d.first_name, d.last_name, s.name").
		Order("bookings_count DESC").
		Limit(5).
		Scan(&topRows)

	for _, tr := range topRows {
		analytics.TopDoctors = append(analytics.TopDoctors, dto.AdminClinicTopDoctorDTO{
			DoctorID:      tr.DoctorID,
			DoctorName:    tr.DoctorName,
			SpecialtyName: tr.SpecialtyName,
			BookingsCount: tr.BookingsCount,
		})
	}

	return analytics, nil
}

func (r *adminRepository) getAppointmentsQuery(ctx context.Context, baseQuery *gorm.DB, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error) {
	filter.Normalize()

	query := baseQuery.WithContext(ctx).Table("appointments a").
		Joins("JOIN users u ON u.id = a.patient_id").
		Joins("JOIN clinics c ON c.id = a.clinic_id").
		Joins("JOIN doctors d ON d.id = a.doctor_id").
		Joins("JOIN clinic_doctors cd ON cd.id = a.clinic_doctor_id").
		Joins("JOIN specialties s ON s.id = cd.specialty_id")

	if filter.ClinicID != "" {
		query = query.Where("a.clinic_id = ?", filter.ClinicID)
	}
	if filter.DoctorID != "" {
		query = query.Where("a.doctor_id = ?", filter.DoctorID)
	}
	if filter.Status != "" {
		query = query.Where("a.status = ?", filter.Status)
	}
	if filter.StartDate != "" {
		query = query.Where("a.appointment_date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("a.appointment_date <= ?", filter.EndDate)
	}
	if filter.PatientSearch != "" {
		s := "%" + strings.ToLower(filter.PatientSearch) + "%"
		query = query.Where("LOWER(u.first_name) LIKE ? OR LOWER(u.last_name) LIKE ? OR u.phone LIKE ?", s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	type ApptRow struct {
		ID                 uuid.UUID
		PatientID          uuid.UUID
		PatientName        string
		PatientPhone       string
		ClinicID           uuid.UUID
		ClinicName         string
		DoctorID           uuid.UUID
		DoctorName         string
		SpecialtyName      string
		AppointmentDate    string
		StartTime          string
		EndTime            string
		Status             string
		ConsultationPrice  float64
		CancellationReason *string
		CreatedAt          time.Time
	}

	var rows []ApptRow
	err := query.Select(`
		a.id, a.patient_id, CONCAT(u.first_name, ' ', u.last_name) as patient_name, u.phone as patient_phone,
		a.clinic_id, c.name as clinic_name,
		a.doctor_id, CONCAT(d.first_name, ' ', d.last_name) as doctor_name,
		s.name as specialty_name,
		a.appointment_date, a.start_time, a.end_time, a.status, a.consultation_price, a.cancellation_reason, a.created_at
	`).
		Order("a.appointment_date DESC, a.start_time DESC").
		Limit(filter.Limit).
		Offset(offset).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	items := make([]dto.AdminAppointmentListItemDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.AdminAppointmentListItemDTO{
			ID:                 row.ID,
			PatientID:          row.PatientID,
			PatientName:        row.PatientName,
			PatientPhone:       row.PatientPhone,
			ClinicID:           row.ClinicID,
			ClinicName:         row.ClinicName,
			DoctorID:           row.DoctorID,
			DoctorName:         row.DoctorName,
			SpecialtyName:      row.SpecialtyName,
			AppointmentDate:    row.AppointmentDate,
			StartTime:          row.StartTime,
			EndTime:            row.EndTime,
			Status:             row.Status,
			ConsultationPrice:  row.ConsultationPrice,
			CancellationReason: row.CancellationReason,
			CreatedAt:          row.CreatedAt,
		})
	}

	return items, total, nil
}
