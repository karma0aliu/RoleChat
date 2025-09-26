package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"rolechat_back/internal/models"
	"rolechat_back/internal/repository"
	"rolechat_back/pkg/config"
	"rolechat_back/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	Create(ctx context.Context, user *models.User) error
	Register(ctx context.Context, email, password, nickname string) (*models.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(refreshToken string) (accessToken string, err error)
}

type userService struct {
	repo repository.UserRepository
	cfg  *config.Config
}

func NewUserService(r repository.UserRepository, cfg *config.Config) UserService {
	return &userService{repo: r, cfg: cfg}
}

func (s *userService) Create(ctx context.Context, user *models.User) error {
	if user.Email == "" {
		return errors.New("email required")
	}
	return s.repo.Create(user)
}

func (s *userService) Register(ctx context.Context, email, password, nickname string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password required")
	}
	if existing, err := s.repo.FindByEmail(email); err == nil && existing != nil {
		return nil, errors.New("email already registered")
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &models.User{Email: email, PasswordHash: hash, Nickname: nickname, Role: "user", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", err
	}
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", errors.New("invalid credentials")
	}
	access, refresh, err := s.issueTokenPair(user, "")
	return access, refresh, err
}

func (s *userService) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	if claims.TokenType != "refresh" {
		return "", errors.New("token is not refresh type")
	}
	stored, err := s.repo.GetRefreshTokenByJTI(claims.ID)
	if err != nil {
		return "", errors.New("refresh token not found")
	}
	if stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return "", errors.New("refresh token expired or revoked")
	}
	sha := sha256.Sum256([]byte(refreshToken))
	providedHash := hex.EncodeToString(sha[:])
	if providedHash != stored.TokenHash {
		return "", errors.New("invalid refresh token")
	}
	accessJTI := uuid.NewString()
	accessExp := time.Duration(s.cfg.JWT.AccessExpiresMins) * time.Minute
	access, genErr := utils.GenerateToken(claims.UserID, claims.Role, claims.Status, "access", s.cfg.JWT.AccessSecret, accessJTI, accessExp)
	if genErr != nil {
		return "", genErr
	}
	return access, nil
}

func (s *userService) issueTokenPair(user *models.User, prevJTI string) (string, string, error) {
	accessJTI := uuid.NewString()
	refreshJTI := uuid.NewString()
	accessExp := time.Duration(s.cfg.JWT.AccessExpiresMins) * time.Minute
	refreshExp := time.Duration(s.cfg.JWT.RefreshExpiresHours) * time.Hour
	access, err := utils.GenerateToken(user.ID, user.Role, user.Status, "access", s.cfg.JWT.AccessSecret, accessJTI, accessExp)
	if err != nil {
		return "", "", err
	}
	refresh, err := utils.GenerateToken(user.ID, user.Role, user.Status, "refresh", s.cfg.JWT.RefreshSecret, refreshJTI, refreshExp)
	if err != nil {
		return "", "", err
	}
	sha := sha256.Sum256([]byte(refresh))
	hash := hex.EncodeToString(sha[:])
	rt := &models.RefreshToken{
		UserID:    user.ID,
		TokenID:   refreshJTI,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(refreshExp),
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveRefreshToken(rt); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}
	if prevJTI != "" {
		_ = s.repo.RevokeRefreshToken(prevJTI, refreshJTI)
	}
	return access, refresh, nil
}
