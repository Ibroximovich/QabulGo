package repositories

import (
	"context"
	"errors"
	"time"

	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("foydalanuvchi topilmadi")
	ErrEmailAlreadyExist = errors.New("ushbu email allaqachon ro‘yxatdan o‘tgan")
	ErrPhoneAlreadyExist = errors.New("ushbu telefon raqam allaqachon ro‘yxatdan o‘tgan")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateEmailVerified(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?)", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) UpdateEmailVerified(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Update("email_verified_at", verifiedAt).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"password_hash": passwordHash,
			"updated_at":    time.Now().UTC(),
		}).Error
}
