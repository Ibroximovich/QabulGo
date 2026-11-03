package services

import (
	"context"
	"testing"
	"time"

	"qabulgo-backend/internal/dto"

	"github.com/google/uuid"
)

type mockAdminRepo struct{}

func (m *mockAdminRepo) GetDashboardSummary(ctx context.Context) (*dto.AdminDashboardSummaryDTO, error) {
	return &dto.AdminDashboardSummaryDTO{
		TotalClinics:          10,
		ApprovedClinics:       8,
		PendingClinics:        2,
		TotalPatients:         50,
		TotalDoctors:          25,
		TotalAppointments:     120,
		TodayAppointments:     5,
		ConfirmedAppointments: 100,
		CancelledAppointments: 10,
		CompletedAppointments: 8,
		NoShowAppointments:    2,
		EstimatedTotalRevenue: 15000000,
	}, nil
}

func (m *mockAdminRepo) GetDashboardAnalytics(ctx context.Context, filter *dto.AdminDashboardAnalyticsFilter) (*dto.AdminDashboardAnalyticsDTO, error) {
	return &dto.AdminDashboardAnalyticsDTO{
		DailyBookings: []dto.DailyMetricDTO{
			{Date: "2026-11-01", Count: 10, Amount: 1500000},
			{Date: "2026-11-02", Count: 15, Amount: 2250000},
		},
		StatusDistribution: []dto.StatusCountDTO{
			{Status: "confirmed", Count: 100},
			{Status: "cancelled", Count: 10},
		},
		TopClinics: []dto.TopClinicMetricDTO{
			{ClinicID: uuid.New(), ClinicName: "Shifo Med", BookingsCount: 45, TotalRevenue: 6750000},
		},
	}, nil
}

func (m *mockAdminRepo) GetClinicDoctors(ctx context.Context, clinicID uuid.UUID) ([]dto.AdminClinicDoctorItemDTO, error) {
	return []dto.AdminClinicDoctorItemDTO{
		{DoctorID: uuid.New(), FirstName: "Ali", LastName: "Valiyev", SpecialtyName: "Kardiolog", ConsultationPrice: 150000, IsActive: true},
	}, nil
}

func (m *mockAdminRepo) GetClinicAppointments(ctx context.Context, clinicID uuid.UUID, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error) {
	return []dto.AdminAppointmentListItemDTO{
		{ID: uuid.New(), PatientName: "Jasur Rahimov", DoctorName: "Ali Valiyev", Status: "confirmed"},
	}, 1, nil
}

func (m *mockAdminRepo) GetClinicAnalytics(ctx context.Context, clinicID uuid.UUID) (*dto.AdminClinicAnalyticsDTO, error) {
	return &dto.AdminClinicAnalyticsDTO{
		TotalAppointments: 45,
		TodayAppointments: 3,
	}, nil
}

func (m *mockAdminRepo) ListPatients(ctx context.Context, filter *dto.AdminPatientFilter) ([]dto.AdminPatientListItemDTO, int64, error) {
	return []dto.AdminPatientListItemDTO{
		{ID: uuid.New(), FirstName: "Sanjar", LastName: "Karimov", Email: "sanjar@example.com", AppointmentsCount: 3, CreatedAt: time.Now()},
	}, 1, nil
}

func (m *mockAdminRepo) GetPatientDetails(ctx context.Context, patientID uuid.UUID) (*dto.AdminPatientDetailDTO, error) {
	return &dto.AdminPatientDetailDTO{
		ID:        patientID,
		FirstName: "Sanjar",
		LastName:  "Karimov",
		Email:     "sanjar@example.com",
	}, nil
}

func (m *mockAdminRepo) ListDoctors(ctx context.Context, filter *dto.AdminDoctorFilter) ([]dto.AdminDoctorListItemDTO, int64, error) {
	return []dto.AdminDoctorListItemDTO{
		{ID: uuid.New(), FirstName: "Ali", LastName: "Valiyev", ClinicsCount: 2, TotalAppointments: 50},
	}, 1, nil
}

func (m *mockAdminRepo) GetDoctorDetails(ctx context.Context, doctorID uuid.UUID) (*dto.AdminDoctorDetailDTO, error) {
	return &dto.AdminDoctorDetailDTO{
		ID:        doctorID,
		FirstName: "Ali",
		LastName:  "Valiyev",
	}, nil
}

func (m *mockAdminRepo) ListAppointments(ctx context.Context, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error) {
	return []dto.AdminAppointmentListItemDTO{
		{ID: uuid.New(), PatientName: "Jasur Rahimov", ClinicName: "Shifo Med", DoctorName: "Ali Valiyev", Status: "confirmed"},
	}, 1, nil
}

func (m *mockAdminRepo) GetAppointmentDetails(ctx context.Context, appointmentID uuid.UUID) (*dto.AdminAppointmentListItemDTO, error) {
	return &dto.AdminAppointmentListItemDTO{
		ID:          appointmentID,
		PatientName: "Jasur Rahimov",
		ClinicName:  "Shifo Med",
	}, nil
}

func TestAdminService_DashboardMetrics(t *testing.T) {
	adminRepo := &mockAdminRepo{}
	svc := NewAdminService(nil, nil, nil, adminRepo)

	summary, err := svc.GetDashboardSummary(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardSummary failed: %v", err)
	}

	if summary.TotalClinics != 10 {
		t.Errorf("Expected 10 total clinics, got %d", summary.TotalClinics)
	}
	if summary.EstimatedTotalRevenue != 15000000 {
		t.Errorf("Expected 15000000 estimated revenue, got %f", summary.EstimatedTotalRevenue)
	}

	analytics, err := svc.GetDashboardAnalytics(context.Background(), &dto.AdminDashboardAnalyticsFilter{Range: "7d"})
	if err != nil {
		t.Fatalf("GetDashboardAnalytics failed: %v", err)
	}
	if len(analytics.DailyBookings) != 2 {
		t.Errorf("Expected 2 daily booking entries, got %d", len(analytics.DailyBookings))
	}
}

func TestAdminService_PatientsAndDoctors(t *testing.T) {
	adminRepo := &mockAdminRepo{}
	svc := NewAdminService(nil, nil, nil, adminRepo)

	patients, totalP, err := svc.ListPatients(context.Background(), &dto.AdminPatientFilter{Page: 1, Limit: 10})
	if err != nil || totalP != 1 || len(patients) != 1 {
		t.Fatalf("ListPatients failed: %v", err)
	}

	doctors, totalD, err := svc.ListDoctors(context.Background(), &dto.AdminDoctorFilter{Page: 1, Limit: 10})
	if err != nil || totalD != 1 || len(doctors) != 1 {
		t.Fatalf("ListDoctors failed: %v", err)
	}

	appts, totalA, err := svc.ListAppointments(context.Background(), &dto.AdminAppointmentFilter{Page: 1, Limit: 10})
	if err != nil || totalA != 1 || len(appts) != 1 {
		t.Fatalf("ListAppointments failed: %v", err)
	}
}

