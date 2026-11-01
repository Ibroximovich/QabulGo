package repositories

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NearbyClinicRaw struct {
	models.Clinic
	DistanceKM float64 `gorm:"column:distance_km"`
}

type PublicSearchRepository interface {
	SearchClinics(ctx context.Context, filter *dto.PublicClinicFilter) ([]models.Clinic, int64, error)
	GetApprovedClinicByID(ctx context.Context, id uuid.UUID) (*models.Clinic, error)
	GetClinicSpecialties(ctx context.Context, clinicID uuid.UUID) ([]models.Specialty, error)
	GetClinicActiveDoctors(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicDoctor, error)
	SearchNearbyClinics(ctx context.Context, filter *dto.NearbyClinicFilter) ([]NearbyClinicRaw, int64, error)
	SearchDoctors(ctx context.Context, filter *dto.PublicDoctorFilter) ([]models.ClinicDoctor, int64, error)
	GetDoctorByIDOrClinicDoctorID(ctx context.Context, targetID uuid.UUID) (*models.Doctor, []models.ClinicDoctor, error)
	GetRegionsCatalog(ctx context.Context) ([]dto.RegionItemDTO, error)
}

type publicSearchRepository struct {
	db *gorm.DB
}

func NewPublicSearchRepository(db *gorm.DB) PublicSearchRepository {
	return &publicSearchRepository{db: db}
}

func (r *publicSearchRepository) SearchClinics(ctx context.Context, filter *dto.PublicClinicFilter) ([]models.Clinic, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Clinic{}).
		Where("status = ?", models.ClinicStatusApproved)

	if filter != nil {
		if strings.TrimSpace(filter.Search) != "" {
			term := "%" + strings.ToLower(strings.TrimSpace(filter.Search)) + "%"
			q = q.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(address) LIKE ?)", term, term, term)
		}
		if strings.TrimSpace(filter.Region) != "" {
			q = q.Where("LOWER(region) = LOWER(?)", strings.TrimSpace(filter.Region))
		}
		if strings.TrimSpace(filter.District) != "" {
			q = q.Where("LOWER(district) = LOWER(?)", strings.TrimSpace(filter.District))
		}
		if filter.SpecialtyID != nil && strings.TrimSpace(*filter.SpecialtyID) != "" {
			q = q.Where("EXISTS (SELECT 1 FROM clinic_doctors cd WHERE cd.clinic_id = clinics.id AND cd.specialty_id = ? AND cd.is_active = true AND cd.deleted_at IS NULL)", strings.TrimSpace(*filter.SpecialtyID))
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := 12
	page := 1
	if filter != nil {
		if filter.Limit > 0 && filter.Limit <= 50 {
			limit = filter.Limit
		} else if filter.Limit > 50 {
			limit = 50
		}
		if filter.Page > 0 {
			page = filter.Page
		}
	}
	offset := (page - 1) * limit

	var clinics []models.Clinic
	err := q.Preload("WorkingHours").
		Order("name ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&clinics).Error
	if err != nil {
		return nil, 0, err
	}

	return clinics, total, nil
}

func (r *publicSearchRepository) GetApprovedClinicByID(ctx context.Context, id uuid.UUID) (*models.Clinic, error) {
	var clinic models.Clinic
	err := r.db.WithContext(ctx).
		Preload("WorkingHours").
		First(&clinic, "id = ? AND status = ?", id, models.ClinicStatusApproved).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClinicNotFound
		}
		return nil, err
	}
	return &clinic, nil
}

func (r *publicSearchRepository) GetClinicSpecialties(ctx context.Context, clinicID uuid.UUID) ([]models.Specialty, error) {
	var specialties []models.Specialty
	err := r.db.WithContext(ctx).
		Distinct("specialties.*").
		Table("specialties").
		Joins("JOIN clinic_doctors cd ON cd.specialty_id = specialties.id").
		Where("cd.clinic_id = ? AND cd.is_active = true AND cd.deleted_at IS NULL", clinicID).
		Order("specialties.name ASC").
		Find(&specialties).Error
	if err != nil {
		return nil, err
	}
	return specialties, nil
}

func (r *publicSearchRepository) GetClinicActiveDoctors(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicDoctor, error) {
	var doctors []models.ClinicDoctor
	err := r.db.WithContext(ctx).
		Model(&models.ClinicDoctor{}).
		Where("clinic_id = ? AND is_active = true AND deleted_at IS NULL", clinicID).
		Preload("Doctor").
		Preload("Specialty").
		Order("consultation_price ASC, created_at ASC").
		Find(&doctors).Error
	if err != nil {
		return nil, err
	}
	return doctors, nil
}

func (r *publicSearchRepository) SearchNearbyClinics(ctx context.Context, filter *dto.NearbyClinicFilter) ([]NearbyClinicRaw, int64, error) {
	lat := filter.Latitude
	lon := filter.Longitude
	radius := filter.RadiusKM
	if radius <= 0 {
		radius = 10
	} else if radius > 100 {
		radius = 100
	}

	limit := 12
	page := 1
	if filter.Limit > 0 && filter.Limit <= 50 {
		limit = filter.Limit
	} else if filter.Limit > 50 {
		limit = 50
	}
	if filter.Page > 0 {
		page = filter.Page
	}
	offset := (page - 1) * limit

	// Bounding box hisoblash (PostgreSQL composite B-Tree indexidan samarali foydalanish uchun)
	degLat := radius / 111.0
	radLat := lat * math.Pi / 180.0
	cosLat := math.Cos(radLat)
	if cosLat < 0.0001 {
		cosLat = 0.0001
	}
	degLon := radius / (111.0 * cosLat)

	minLat := lat - degLat
	maxLat := lat + degLat
	minLon := lon - degLon
	maxLon := lon + degLon

	// Haversine formulasi (km)
	haversineSql := "(6371.0 * acos(least(1.0, greatest(-1.0, cos(radians(?)) * cos(radians(clinics.latitude)) * cos(radians(clinics.longitude) - radians(?)) + sin(radians(?)) * sin(radians(clinics.latitude))))))"

	q := r.db.WithContext(ctx).Table("clinics").
		Where("clinics.status = ?", models.ClinicStatusApproved).
		Where("clinics.latitude IS NOT NULL AND clinics.longitude IS NOT NULL").
		Where("clinics.latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("clinics.longitude BETWEEN ? AND ?", minLon, maxLon)

	if filter.SpecialtyID != nil && strings.TrimSpace(*filter.SpecialtyID) != "" {
		q = q.Where("EXISTS (SELECT 1 FROM clinic_doctors cd WHERE cd.clinic_id = clinics.id AND cd.specialty_id = ? AND cd.is_active = true AND cd.deleted_at IS NULL)", strings.TrimSpace(*filter.SpecialtyID))
	}

	// Exact radius filter
	q = q.Where(haversineSql+" <= ?", lat, lon, lat, radius)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSql := fmt.Sprintf("clinics.*, %s AS distance_km", haversineSql)
	var items []NearbyClinicRaw
	err := q.Select(selectSql, lat, lon, lat).
		Order("distance_km ASC, clinics.name ASC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *publicSearchRepository) SearchDoctors(ctx context.Context, filter *dto.PublicDoctorFilter) ([]models.ClinicDoctor, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.ClinicDoctor{}).
		Joins("JOIN clinics ON clinics.id = clinic_doctors.clinic_id").
		Joins("JOIN doctors ON doctors.id = clinic_doctors.doctor_id").
		Where("clinics.status = ?", models.ClinicStatusApproved).
		Where("clinic_doctors.is_active = true").
		Where("clinic_doctors.deleted_at IS NULL")

	if filter != nil {
		if strings.TrimSpace(filter.Search) != "" {
			term := "%" + strings.ToLower(strings.TrimSpace(filter.Search)) + "%"
			q = q.Where("(LOWER(doctors.first_name) LIKE ? OR LOWER(doctors.last_name) LIKE ? OR LOWER(doctors.first_name || ' ' || doctors.last_name) LIKE ? OR LOWER(doctors.last_name || ' ' || doctors.first_name) LIKE ? OR LOWER(doctors.bio) LIKE ?)", term, term, term, term, term)
		}
		if strings.TrimSpace(filter.Region) != "" {
			q = q.Where("LOWER(clinics.region) = LOWER(?)", strings.TrimSpace(filter.Region))
		}
		if strings.TrimSpace(filter.District) != "" {
			q = q.Where("LOWER(clinics.district) = LOWER(?)", strings.TrimSpace(filter.District))
		}
		if filter.SpecialtyID != nil && strings.TrimSpace(*filter.SpecialtyID) != "" {
			q = q.Where("clinic_doctors.specialty_id = ?", strings.TrimSpace(*filter.SpecialtyID))
		}
		if filter.ClinicID != nil && strings.TrimSpace(*filter.ClinicID) != "" {
			q = q.Where("clinic_doctors.clinic_id = ?", strings.TrimSpace(*filter.ClinicID))
		}
		if filter.MinPrice != nil && *filter.MinPrice >= 0 {
			q = q.Where("clinic_doctors.consultation_price >= ?", *filter.MinPrice)
		}
		if filter.MaxPrice != nil && *filter.MaxPrice >= 0 {
			q = q.Where("clinic_doctors.consultation_price <= ?", *filter.MaxPrice)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := 12
	page := 1
	sortOrder := "name_asc"
	if filter != nil {
		if filter.Limit > 0 && filter.Limit <= 50 {
			limit = filter.Limit
		} else if filter.Limit > 50 {
			limit = 50
		}
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Sort != "" {
			sortOrder = filter.Sort
		}
	}
	offset := (page - 1) * limit

	switch sortOrder {
	case "price_asc":
		q = q.Order("clinic_doctors.consultation_price ASC, doctors.last_name ASC")
	case "price_desc":
		q = q.Order("clinic_doctors.consultation_price DESC, doctors.last_name ASC")
	case "experience_desc":
		q = q.Order("doctors.experience_years DESC, clinic_doctors.consultation_price ASC")
	default:
		q = q.Order("doctors.last_name ASC, doctors.first_name ASC")
	}

	var items []models.ClinicDoctor
	err := q.Preload("Doctor").
		Preload("Specialty").
		Preload("Clinic").
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *publicSearchRepository) GetDoctorByIDOrClinicDoctorID(ctx context.Context, targetID uuid.UUID) (*models.Doctor, []models.ClinicDoctor, error) {
	var doctor models.Doctor
	// 1. Avval doctors jadvalida qidirib ko'ramiz
	err := r.db.WithContext(ctx).First(&doctor, "id = ?", targetID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 2. Agar topilmasa, clinic_doctor id bo'lishi mumkin
			var cd models.ClinicDoctor
			errCd := r.db.WithContext(ctx).Preload("Doctor").First(&cd, "id = ?", targetID).Error
			if errCd != nil {
				return nil, nil, ErrDoctorNotFound
			}
			if cd.Doctor != nil {
				doctor = *cd.Doctor
			} else {
				return nil, nil, ErrDoctorNotFound
			}
		} else {
			return nil, nil, err
		}
	}

	// 3. Ushbu shifokorning barcha approved klinikalardagi faol ish joylarini yuklaymiz
	var workplaces []models.ClinicDoctor
	err = r.db.WithContext(ctx).
		Model(&models.ClinicDoctor{}).
		Joins("JOIN clinics ON clinics.id = clinic_doctors.clinic_id").
		Where("clinic_doctors.doctor_id = ?", doctor.ID).
		Where("clinics.status = ?", models.ClinicStatusApproved).
		Where("clinic_doctors.is_active = true").
		Where("clinic_doctors.deleted_at IS NULL").
		Preload("Clinic").
		Preload("Specialty").
		Order("clinic_doctors.consultation_price ASC").
		Find(&workplaces).Error
	if err != nil {
		return nil, nil, err
	}

	if len(workplaces) == 0 {
		return nil, nil, ErrDoctorNotFound
	}

	return &doctor, workplaces, nil
}

func (r *publicSearchRepository) GetRegionsCatalog(ctx context.Context) ([]dto.RegionItemDTO, error) {
	type RegionRow struct {
		Region   string `gorm:"column:region"`
		District string `gorm:"column:district"`
	}

	var rows []RegionRow
	err := r.db.WithContext(ctx).
		Table("clinics").
		Distinct("region", "district").
		Where("status = ?", models.ClinicStatusApproved).
		Order("region ASC, district ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	regionMap := make(map[string][]string)
	var orderedRegions []string
	for _, row := range rows {
		reg := strings.TrimSpace(row.Region)
		dist := strings.TrimSpace(row.District)
		if reg == "" {
			continue
		}
		if _, exists := regionMap[reg]; !exists {
			regionMap[reg] = []string{}
			orderedRegions = append(orderedRegions, reg)
		}
		if dist != "" {
			regionMap[reg] = append(regionMap[reg], dist)
		}
	}

	var result []dto.RegionItemDTO
	for _, reg := range orderedRegions {
		result = append(result, dto.RegionItemDTO{
			Region:    reg,
			Districts: regionMap[reg],
		})
	}

	return result, nil
}

