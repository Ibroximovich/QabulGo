package services

import (
	"context"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/repositories"
)

type SpecialtyService interface {
	ListSpecialties(ctx context.Context) ([]dto.SpecialtyDTO, error)
	SeedIfEmpty(ctx context.Context) error
}

type specialtyService struct {
	specialtyRepo repositories.SpecialtyRepository
}

func NewSpecialtyService(specialtyRepo repositories.SpecialtyRepository) SpecialtyService {
	return &specialtyService{specialtyRepo: specialtyRepo}
}

func (s *specialtyService) ListSpecialties(ctx context.Context) ([]dto.SpecialtyDTO, error) {
	list, err := s.specialtyRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]dto.SpecialtyDTO, 0, len(list))
	for _, item := range list {
		res = append(res, dto.SpecialtyDTO{
			ID:      item.ID,
			Name:    item.Name,
			Slug:    item.Slug,
			IconURL: item.IconURL,
		})
	}
	return res, nil
}

func (s *specialtyService) SeedIfEmpty(ctx context.Context) error {
	return s.specialtyRepo.SeedIfEmpty(ctx)
}

