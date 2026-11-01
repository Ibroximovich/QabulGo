package repositories

import (
	"context"
	"errors"

	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSpecialtyNotFound = errors.New("mutaxassislik topilmadi")
)

type SpecialtyRepository interface {
	ListAll(ctx context.Context) ([]models.Specialty, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Specialty, error)
	GetBySlug(ctx context.Context, slug string) (*models.Specialty, error)
	SeedIfEmpty(ctx context.Context) error
}

type specialtyRepository struct {
	db *gorm.DB
}

func NewSpecialtyRepository(db *gorm.DB) SpecialtyRepository {
	return &specialtyRepository{db: db}
}

func (r *specialtyRepository) ListAll(ctx context.Context) ([]models.Specialty, error) {
	var list []models.Specialty
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *specialtyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Specialty, error) {
	var s models.Specialty
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSpecialtyNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *specialtyRepository) GetBySlug(ctx context.Context, slug string) (*models.Specialty, error) {
	var s models.Specialty
	err := r.db.WithContext(ctx).First(&s, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSpecialtyNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *specialtyRepository) SeedIfEmpty(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Specialty{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaultSpecialties := []models.Specialty{
		{ID: uuid.New(), Name: "Terapevt", Slug: "terapevt"},
		{ID: uuid.New(), Name: "Nevrolog", Slug: "nevrolog"},
		{ID: uuid.New(), Name: "Kardiolog", Slug: "kardiolog"},
		{ID: uuid.New(), Name: "Urolog", Slug: "urolog"},
		{ID: uuid.New(), Name: "Pediatr", Slug: "pediatr"},
		{ID: uuid.New(), Name: "Dermatolog", Slug: "dermatolog"},
		{ID: uuid.New(), Name: "Ginekolog", Slug: "ginekolog"},
		{ID: uuid.New(), Name: "Oftalmolog", Slug: "oftalmolog"},
		{ID: uuid.New(), Name: "LOR (Otorinolaringolog)", Slug: "lor"},
		{ID: uuid.New(), Name: "Stomatolog", Slug: "stomatolog"},
		{ID: uuid.New(), Name: "Xirurg", Slug: "xirurg"},
		{ID: uuid.New(), Name: "Endokrinolog", Slug: "endokrinolog"},
		{ID: uuid.New(), Name: "Gastroenterolog", Slug: "gastroenterolog"},
		{ID: uuid.New(), Name: "Ortoped-Travmatolog", Slug: "ortoped-travmatolog"},
		{ID: uuid.New(), Name: "Allergolog", Slug: "allergolog"},
	}

	return r.db.WithContext(ctx).Create(&defaultSpecialties).Error
}

