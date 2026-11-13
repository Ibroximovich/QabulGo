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
	defaultSpecialties := []models.Specialty{
		{Name: "Terapevt", Slug: "terapevt"},
		{Name: "Nevrolog", Slug: "nevrolog"},
		{Name: "Kardiolog", Slug: "kardiolog"},
		{Name: "Urolog", Slug: "urolog"},
		{Name: "Pediatr", Slug: "pediatr"},
		{Name: "Dermatolog", Slug: "dermatolog"},
		{Name: "Ginekolog", Slug: "ginekolog"},
		{Name: "Oftalmolog", Slug: "oftalmolog"},
		{Name: "LOR (Otorinolaringolog)", Slug: "lor"},
		{Name: "Stomatolog", Slug: "stomatolog"},
		{Name: "Xirurg", Slug: "xirurg"},
		{Name: "Endokrinolog", Slug: "endokrinolog"},
		{Name: "Gastroenterolog", Slug: "gastroenterolog"},
		{Name: "Ortoped-Travmatolog", Slug: "ortoped-travmatolog"},
		{Name: "Allergolog", Slug: "allergolog"},
		// Yangi 21 ta shifokor mutaxassisliklari
		{Name: "Onkolog", Slug: "onkolog"},
		{Name: "Gematolog", Slug: "gematolog"},
		{Name: "Pulmonolog", Slug: "pulmonolog"},
		{Name: "Nefrolog", Slug: "nefrolog"},
		{Name: "Revmatolog", Slug: "revmatolog"},
		{Name: "Infeksionist", Slug: "infeksionist"},
		{Name: "Psixiatr", Slug: "psixiatr"},
		{Name: "Psixoterapevt", Slug: "psixoterapevt"},
		{Name: "Neyroxirurg", Slug: "neyroxirurg"},
		{Name: "Anesteziolog-reanimatolog", Slug: "anesteziolog-reanimatolog"},
		{Name: "Mammolog", Slug: "mammolog"},
		{Name: "Proktolog", Slug: "proktolog"},
		{Name: "Androlog", Slug: "androlog"},
		{Name: "Gepatolog", Slug: "gepatolog"},
		{Name: "Angioxirurg", Slug: "angioxirurg"},
		{Name: "Onkoginekolog", Slug: "onkoginekolog"},
		{Name: "Bolalar nevrologi", Slug: "bolalar-nevrologi"},
		{Name: "Bolalar jarrohi", Slug: "bolalar-jarrohi"},
		{Name: "Neonatolog", Slug: "neonatolog"},
		{Name: "Immunolog", Slug: "immunolog"},
		{Name: "Parazitolog", Slug: "parazitolog"},
	}

	for _, spec := range defaultSpecialties {
		var exists int64
		if err := r.db.WithContext(ctx).Model(&models.Specialty{}).Where("slug = ?", spec.Slug).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			spec.ID = uuid.New()
			if err := r.db.WithContext(ctx).Create(&spec).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

