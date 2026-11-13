package services

import (
	"context"
	"testing"

	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"

	"github.com/google/uuid"
)

type mockSpecialtyRepo struct {
	items []models.Specialty
}

func (m *mockSpecialtyRepo) ListAll(ctx context.Context) ([]models.Specialty, error) {
	return m.items, nil
}

func (m *mockSpecialtyRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Specialty, error) {
	for _, item := range m.items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, repositories.ErrSpecialtyNotFound
}

func (m *mockSpecialtyRepo) GetBySlug(ctx context.Context, slug string) (*models.Specialty, error) {
	for _, item := range m.items {
		if item.Slug == slug {
			return &item, nil
		}
	}
	return nil, repositories.ErrSpecialtyNotFound
}

func (m *mockSpecialtyRepo) SeedIfEmpty(ctx context.Context) error {
	return nil
}

func TestSpecialtyService_ListSpecialties(t *testing.T) {
	expectedNames := []string{
		// 15 existing
		"Allergolog", "Dermatolog", "Endokrinolog", "Gastroenterolog", "Ginekolog",
		"Kardiolog", "LOR (Otorinolaringolog)", "Nevrolog", "Oftalmolog", "Ortoped-Travmatolog",
		"Pediatr", "Stomatolog", "Terapevt", "Urolog", "Xirurg",
		// 21 new
		"Onkolog", "Gematolog", "Pulmonolog", "Nefrolog", "Revmatolog",
		"Infeksionist", "Psixiatr", "Psixoterapevt", "Neyroxirurg", "Anesteziolog-reanimatolog",
		"Mammolog", "Proktolog", "Androlog", "Gepatolog", "Angioxirurg",
		"Onkoginekolog", "Bolalar nevrologi", "Bolalar jarrohi", "Neonatolog", "Immunolog",
		"Parazitolog",
	}

	items := make([]models.Specialty, len(expectedNames))
	for i, name := range expectedNames {
		items[i] = models.Specialty{
			ID:   uuid.New(),
			Name: name,
			Slug: name,
		}
	}

	repo := &mockSpecialtyRepo{items: items}
	svc := NewSpecialtyService(repo)

	dtos, err := svc.ListSpecialties(context.Background())
	if err != nil {
		t.Fatalf("ListSpecialties failed: %v", err)
	}

	if len(dtos) != 36 {
		t.Fatalf("Expected 36 specialties, got %d", len(dtos))
	}

	// Verify Parazitolog exists
	foundParazitolog := false
	for _, dto := range dtos {
		if dto.Name == "Parazitolog" {
			foundParazitolog = true
			break
		}
	}
	if !foundParazitolog {
		t.Errorf("Expected 'Parazitolog' to be present in specialties list")
	}
}

