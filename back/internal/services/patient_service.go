package services

import (
	"context"
	"errors"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

type PatientService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.PatientProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdatePatientProfileRequest) (*dto.PatientProfileResponse, error)
}

type patientService struct {
	userRepo repositories.UserRepository
}

func NewPatientService(userRepo repositories.UserRepository) PatientService {
	return &patientService{userRepo: userRepo}
}

func (s *patientService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.PatientProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Role != models.RolePatient {
		return nil, errors.New("ruxsat etilmagan profil turi")
	}

	return &dto.PatientProfileResponse{
		ID:            user.ID,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Phone:         user.Phone,
		Role:          string(user.Role),
		EmailVerified: user.IsEmailVerified(),
		CreatedAt:     user.CreatedAt,
	}, nil
}

func (s *patientService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdatePatientProfileRequest) (*dto.PatientProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	firstName := utils.SanitizeString(req.FirstName)
	lastName := utils.SanitizeString(req.LastName)
	if len(firstName) < 2 || len(firstName) > 50 || len(lastName) < 2 || len(lastName) > 50 {
		return nil, ErrNameLength
	}

	phone := utils.SanitizeString(req.Phone)
	if !utils.ValidateUzbekistanPhone(phone) {
		return nil, ErrInvalidPhoneFormat
	}

	// Agar telefon o'zgartirilgan bo'lsa, boshqa foydalanuvchida bor-yo'qligini tekshirish
	if phone != user.Phone {
		existingUser, err := s.userRepo.GetByPhone(ctx, phone)
		if err == nil && existingUser.ID != user.ID {
			return nil, repositories.ErrPhoneAlreadyExist
		}
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Phone = phone
	user.UpdatedAt = time.Now().UTC()

	// Email va Role o'zgarmasligi kafolatlanadi
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &dto.PatientProfileResponse{
		ID:            user.ID,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Phone:         user.Phone,
		Role:          string(user.Role),
		EmailVerified: user.IsEmailVerified(),
		CreatedAt:     user.CreatedAt,
	}, nil
}
