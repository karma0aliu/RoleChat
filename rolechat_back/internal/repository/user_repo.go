package repository

import (
	"context"

	"rolechat_back/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	SaveRefreshToken(t *models.RefreshToken) error
	GetRefreshTokenByJTI(jti string) (*models.RefreshToken, error)
	RevokeRefreshToken(jti, replacedBy string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.WithContext(context.Background()).Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(context.Background()).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) SaveRefreshToken(t *models.RefreshToken) error {
	return r.db.WithContext(context.Background()).Create(t).Error
}

func (r *userRepository) GetRefreshTokenByJTI(jti string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	if err := r.db.WithContext(context.Background()).Where("token_id = ?", jti).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *userRepository) RevokeRefreshToken(jti, replacedBy string) error {
	return r.db.WithContext(context.Background()).Model(&models.RefreshToken{}).Where("token_id = ? AND revoked = false", jti).Updates(map[string]any{"revoked": true, "replaced_by": replacedBy}).Error
}

