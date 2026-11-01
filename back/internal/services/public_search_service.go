package services

import (
	"context"
	"errors"
	"math"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"

	"github.com/google/uuid"
)

var (
	ErrInvalidCoordinates = errors.New("noto‘g‘ri geografik koordinatalar: latitude (-90..90) va longitude (-180..180) bo‘lishi shart")
	ErrInvalidRadius      = errors.New("noto‘g‘ri qidiruv radiusi: radius 0.1 km dan 100 km gacha bo‘lishi kerak")
)

type PublicSearchService interface {
	SearchClinics(ctx context.Context, filter *dto.PublicClinicFilter) ([]dto.PublicClinicItemDTO, int64, error)
	GetClinicDetails(ctx context.Context, clinicID uuid.UUID) (*dto.ClinicDetailDTO, error)
	GetClinicDoctors(ctx context.Context, clinicID uuid.UUID) ([]dto.ClinicDoctorItemDTO, error)
	SearchNearbyClinics(ctx context.Context, filter *dto.NearbyClinicFilter) ([]dto.NearbyClinicItemDTO, int64, error)
	SearchDoctors(ctx context.Context, filter *dto.PublicDoctorFilter) ([]dto.PublicDoctorItemDTO, int64, error)
	GetDoctorDetails(ctx context.Context, targetID uuid.UUID) (*dto.PublicDoctorDetailDTO, error)
	GetRegionsCatalog(ctx context.Context) ([]dto.RegionItemDTO, error)
}

type publicSearchService struct {
	repo       repositories.PublicSearchRepository
	doctorRepo repositories.DoctorRepository
}

func NewPublicSearchService(repo repositories.PublicSearchRepository, doctorRepo repositories.DoctorRepository) PublicSearchService {
	return &publicSearchService{
		repo:       repo,
		doctorRepo: doctorRepo,
	}
}

func (s *publicSearchService) SearchClinics(ctx context.Context, filter *dto.PublicClinicFilter) ([]dto.PublicClinicItemDTO, int64, error) {
	clinics, total, err := s.repo.SearchClinics(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.PublicClinicItemDTO, 0, len(clinics))
	for _, c := range clinics {
		item := s.toPublicClinicItemDTO(ctx, &c)
		result = append(result, item)
	}

	return result, total, nil
}

func (s *publicSearchService) GetClinicDetails(ctx context.Context, clinicID uuid.UUID) (*dto.ClinicDetailDTO, error) {
	clinic, err := s.repo.GetApprovedClinicByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	specs, _ := s.repo.GetClinicSpecialties(ctx, clinicID)
	specDTOs := make([]dto.SpecialtyDTO, 0, len(specs))
	for _, spec := range specs {
		specDTOs = append(specDTOs, dto.SpecialtyDTO{
			ID:      spec.ID,
			Name:    spec.Name,
			Slug:    spec.Slug,
			IconURL: spec.IconURL,
		})
	}

	activeDocs, _ := s.repo.GetClinicActiveDoctors(ctx, clinicID)
	docDTOs := make([]dto.ClinicDoctorItemDTO, 0, len(activeDocs))
	for _, cd := range activeDocs {
		var spec *dto.SpecialtyDTO
		if cd.Specialty != nil {
			spec = &dto.SpecialtyDTO{
				ID:      cd.Specialty.ID,
				Name:    cd.Specialty.Name,
				Slug:    cd.Specialty.Slug,
				IconURL: cd.Specialty.IconURL,
			}
		}

		firstName := ""
		lastName := ""
		bio := ""
		exp := 0
		photo := ""
		if cd.Doctor != nil {
			firstName = cd.Doctor.FirstName
			lastName = cd.Doctor.LastName
			bio = cd.Doctor.Bio
			exp = cd.Doctor.ExperienceYears
			photo = cd.Doctor.PhotoURL
		}

		docDTOs = append(docDTOs, dto.ClinicDoctorItemDTO{
			ID:                         cd.ID,
			DoctorID:                   cd.DoctorID,
			ClinicID:                   cd.ClinicID,
			FirstName:                  firstName,
			LastName:                   lastName,
			Bio:                        bio,
			ExperienceYears:            exp,
			PhotoURL:                   photo,
			Specialty:                  spec,
			ConsultationPrice:          cd.ConsultationPrice,
			AppointmentDurationMinutes: cd.AppointmentDurationMinutes,
			IsActive:                   cd.IsActive,
			CreatedAt:                  cd.CreatedAt,
			UpdatedAt:                  cd.UpdatedAt,
		})
	}

	whItems := make([]dto.WorkingHourItem, 0, len(clinic.WorkingHours))
	for _, h := range clinic.WorkingHours {
		whItems = append(whItems, dto.WorkingHourItem{
			Weekday:     h.Weekday,
			OpeningTime: h.OpeningTime,
			ClosingTime: h.ClosingTime,
			IsClosed:    h.IsClosed,
		})
	}

	return &dto.ClinicDetailDTO{
		ID:            clinic.ID,
		Name:          clinic.Name,
		Description:   clinic.Description,
		Phone:         clinic.Phone,
		Email:         clinic.Email,
		Region:        clinic.Region,
		District:      clinic.District,
		Address:       clinic.Address,
		LocationURL:   clinic.LocationURL,
		Latitude:      clinic.Latitude,
		Longitude:     clinic.Longitude,
		LogoURL:       clinic.LogoURL,
		CoverImageURL: clinic.CoverImageURL,
		Status:        string(clinic.Status),
		WorkingHours:  whItems,
		DoctorsCount:  len(activeDocs),
		Specialties:   specDTOs,
		Doctors:       docDTOs,
		CreatedAt:     clinic.CreatedAt,
		UpdatedAt:     clinic.UpdatedAt,
	}, nil
}

func (s *publicSearchService) GetClinicDoctors(ctx context.Context, clinicID uuid.UUID) ([]dto.ClinicDoctorItemDTO, error) {
	// 1. Klinika approved ekanini tekshiramiz
	_, err := s.repo.GetApprovedClinicByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	activeDocs, err := s.repo.GetClinicActiveDoctors(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	docDTOs := make([]dto.ClinicDoctorItemDTO, 0, len(activeDocs))
	for _, cd := range activeDocs {
		var spec *dto.SpecialtyDTO
		if cd.Specialty != nil {
			spec = &dto.SpecialtyDTO{
				ID:      cd.Specialty.ID,
				Name:    cd.Specialty.Name,
				Slug:    cd.Specialty.Slug,
				IconURL: cd.Specialty.IconURL,
			}
		}

		firstName := ""
		lastName := ""
		bio := ""
		exp := 0
		photo := ""
		if cd.Doctor != nil {
			firstName = cd.Doctor.FirstName
			lastName = cd.Doctor.LastName
			bio = cd.Doctor.Bio
			exp = cd.Doctor.ExperienceYears
			photo = cd.Doctor.PhotoURL
		}

		docDTOs = append(docDTOs, dto.ClinicDoctorItemDTO{
			ID:                         cd.ID,
			DoctorID:                   cd.DoctorID,
			ClinicID:                   cd.ClinicID,
			FirstName:                  firstName,
			LastName:                   lastName,
			Bio:                        bio,
			ExperienceYears:            exp,
			PhotoURL:                   photo,
			Specialty:                  spec,
			ConsultationPrice:          cd.ConsultationPrice,
			AppointmentDurationMinutes: cd.AppointmentDurationMinutes,
			IsActive:                   cd.IsActive,
			CreatedAt:                  cd.CreatedAt,
			UpdatedAt:                  cd.UpdatedAt,
		})
	}

	return docDTOs, nil
}

func (s *publicSearchService) SearchNearbyClinics(ctx context.Context, filter *dto.NearbyClinicFilter) ([]dto.NearbyClinicItemDTO, int64, error) {
	if filter == nil {
		return nil, 0, ErrInvalidCoordinates
	}

	// Geolocation validation
	if filter.Latitude < -90.0 || filter.Latitude > 90.0 || filter.Longitude < -180.0 || filter.Longitude > 180.0 {
		return nil, 0, ErrInvalidCoordinates
	}

	if filter.RadiusKM <= 0 {
		filter.RadiusKM = 10.0
	} else if filter.RadiusKM > 100.0 {
		filter.RadiusKM = 100.0
	}

	rawItems, total, err := s.repo.SearchNearbyClinics(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.NearbyClinicItemDTO, 0, len(rawItems))
	for _, it := range rawItems {
		specs, _ := s.repo.GetClinicSpecialties(ctx, it.ID)
		specDTOs := make([]dto.SpecialtyDTO, 0, len(specs))
		for _, sp := range specs {
			specDTOs = append(specDTOs, dto.SpecialtyDTO{
				ID:      sp.ID,
				Name:    sp.Name,
				Slug:    sp.Slug,
				IconURL: sp.IconURL,
			})
		}

		docCount, _ := s.doctorRepo.CountDoctorsByClinic(ctx, it.ID)

		lat := 0.0
		if it.Latitude != nil {
			lat = *it.Latitude
		}
		lon := 0.0
		if it.Longitude != nil {
			lon = *it.Longitude
		}

		// Masofani 2 xona aniqlikda yaxlitlash (masalan 3.45 km)
		distanceRounded := math.Round(it.DistanceKM*100) / 100

		result = append(result, dto.NearbyClinicItemDTO{
			ID:            it.ID,
			Name:          it.Name,
			Description:   it.Description,
			Phone:         it.Phone,
			Email:         it.Email,
			Region:        it.Region,
			District:      it.District,
			Address:       it.Address,
			LocationURL:   it.LocationURL,
			LogoURL:       it.LogoURL,
			CoverImageURL: it.CoverImageURL,
			Latitude:      lat,
			Longitude:     lon,
			DistanceKM:    distanceRounded,
			DoctorsCount:  docCount,
			Specialties:   specDTOs,
		})
	}

	return result, total, nil
}

func (s *publicSearchService) SearchDoctors(ctx context.Context, filter *dto.PublicDoctorFilter) ([]dto.PublicDoctorItemDTO, int64, error) {
	clinicDoctors, total, err := s.repo.SearchDoctors(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.PublicDoctorItemDTO, 0, len(clinicDoctors))
	for _, cd := range clinicDoctors {
		var specDTO *dto.SpecialtyDTO
		if cd.Specialty != nil {
			specDTO = &dto.SpecialtyDTO{
				ID:      cd.Specialty.ID,
				Name:    cd.Specialty.Name,
				Slug:    cd.Specialty.Slug,
				IconURL: cd.Specialty.IconURL,
			}
		}

		var clinicDTO dto.PublicDoctorClinicDTO
		if cd.Clinic != nil {
			clinicDTO = dto.PublicDoctorClinicDTO{
				ID:          cd.Clinic.ID,
				Name:        cd.Clinic.Name,
				Region:      cd.Clinic.Region,
				District:    cd.Clinic.District,
				Address:     cd.Clinic.Address,
				Phone:       cd.Clinic.Phone,
				LocationURL: cd.Clinic.LocationURL,
				LogoURL:     cd.Clinic.LogoURL,
				Latitude:    cd.Clinic.Latitude,
				Longitude:   cd.Clinic.Longitude,
			}
		}

		firstName := ""
		lastName := ""
		photo := ""
		bio := ""
		exp := 0
		if cd.Doctor != nil {
			firstName = cd.Doctor.FirstName
			lastName = cd.Doctor.LastName
			photo = cd.Doctor.PhotoURL
			bio = cd.Doctor.Bio
			exp = cd.Doctor.ExperienceYears
		}

		result = append(result, dto.PublicDoctorItemDTO{
			ClinicDoctorID:             cd.ID,
			DoctorID:                   cd.DoctorID,
			FirstName:                  firstName,
			LastName:                   lastName,
			PhotoURL:                   photo,
			Bio:                        bio,
			ExperienceYears:            exp,
			Specialty:                  specDTO,
			ConsultationPrice:          cd.ConsultationPrice,
			AppointmentDurationMinutes: cd.AppointmentDurationMinutes,
			Clinic:                     clinicDTO,
		})
	}

	return result, total, nil
}

func (s *publicSearchService) GetDoctorDetails(ctx context.Context, targetID uuid.UUID) (*dto.PublicDoctorDetailDTO, error) {
	doctor, workplaces, err := s.repo.GetDoctorByIDOrClinicDoctorID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	workplaceDTOs := make([]dto.DoctorClinicWorkplaceDTO, 0, len(workplaces))
	var primarySpec *dto.SpecialtyDTO

	for _, wp := range workplaces {
		var specDTO *dto.SpecialtyDTO
		if wp.Specialty != nil {
			specDTO = &dto.SpecialtyDTO{
				ID:      wp.Specialty.ID,
				Name:    wp.Specialty.Name,
				Slug:    wp.Specialty.Slug,
				IconURL: wp.Specialty.IconURL,
			}
			if primarySpec == nil {
				primarySpec = specDTO
			}
		}

		clinicName := ""
		region := ""
		district := ""
		address := ""
		phone := ""
		locationURL := ""
		logoURL := ""
		var lat, lon *float64

		if wp.Clinic != nil {
			clinicName = wp.Clinic.Name
			region = wp.Clinic.Region
			district = wp.Clinic.District
			address = wp.Clinic.Address
			phone = wp.Clinic.Phone
			locationURL = wp.Clinic.LocationURL
			logoURL = wp.Clinic.LogoURL
			lat = wp.Clinic.Latitude
			lon = wp.Clinic.Longitude
		}

		workplaceDTOs = append(workplaceDTOs, dto.DoctorClinicWorkplaceDTO{
			ClinicDoctorID:             wp.ID,
			ClinicID:                   wp.ClinicID,
			ClinicName:                 clinicName,
			Region:                     region,
			District:                   district,
			Address:                    address,
			Phone:                      phone,
			LocationURL:                locationURL,
			LogoURL:                    logoURL,
			Latitude:                   lat,
			Longitude:                  lon,
			Specialty:                  specDTO,
			ConsultationPrice:          wp.ConsultationPrice,
			AppointmentDurationMinutes: wp.AppointmentDurationMinutes,
		})
	}

	return &dto.PublicDoctorDetailDTO{
		DoctorID:         doctor.ID,
		FirstName:        doctor.FirstName,
		LastName:         doctor.LastName,
		PhotoURL:         doctor.PhotoURL,
		Bio:              doctor.Bio,
		ExperienceYears:  doctor.ExperienceYears,
		PrimarySpecialty: primarySpec,
		Workplaces:       workplaceDTOs,
	}, nil
}

func (s *publicSearchService) GetRegionsCatalog(ctx context.Context) ([]dto.RegionItemDTO, error) {
	return s.repo.GetRegionsCatalog(ctx)
}

func (s *publicSearchService) toPublicClinicItemDTO(ctx context.Context, c *models.Clinic) dto.PublicClinicItemDTO {
	whItems := make([]dto.WorkingHourItem, 0, len(c.WorkingHours))
	for _, h := range c.WorkingHours {
		whItems = append(whItems, dto.WorkingHourItem{
			Weekday:     h.Weekday,
			OpeningTime: h.OpeningTime,
			ClosingTime: h.ClosingTime,
			IsClosed:    h.IsClosed,
		})
	}

	specs, _ := s.repo.GetClinicSpecialties(ctx, c.ID)
	specDTOs := make([]dto.SpecialtyDTO, 0, len(specs))
	for _, sp := range specs {
		specDTOs = append(specDTOs, dto.SpecialtyDTO{
			ID:      sp.ID,
			Name:    sp.Name,
			Slug:    sp.Slug,
			IconURL: sp.IconURL,
		})
	}

	docCount, _ := s.doctorRepo.CountDoctorsByClinic(ctx, c.ID)

	return dto.PublicClinicItemDTO{
		ID:            c.ID,
		Name:          c.Name,
		Description:   c.Description,
		Phone:         c.Phone,
		Email:         c.Email,
		Region:        c.Region,
		District:      c.District,
		Address:       c.Address,
		LocationURL:   c.LocationURL,
		LogoURL:       c.LogoURL,
		CoverImageURL: c.CoverImageURL,
		Latitude:      c.Latitude,
		Longitude:     c.Longitude,
		WorkingHours:  whItems,
		DoctorsCount:  docCount,
		Specialties:   specDTOs,
		CreatedAt:     c.CreatedAt,
	}
}

