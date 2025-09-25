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
	RefreshTokens(refreshToken string) (accessToken, newRefreshToken string, err error)
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
	// Check if exists
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

func (s *userService) RefreshTokens(refreshToken string) (string, string, error) {
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	if claims.TokenType != "refresh" {
		return "", "", errors.New("token is not refresh type")
	}
	// Load stored token
	stored, err := s.repo.GetRefreshTokenByJTI(claims.ID)
	if err != nil {
		return "", "", errors.New("refresh token not found")
	}
	if stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return "", "", errors.New("refresh token expired or revoked")
	}
	// Rotate: revoke old & issue new
	user := &models.User{ID: claims.UserID, Role: claims.Role, Status: claims.Status}
	access, newRefresh, err := s.issueTokenPair(user, stored.TokenID)
	return access, newRefresh, err
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
	// hash refresh token for storage
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
		_ = s.repo.RevokeRefreshToken(prevJTI, refreshJTI) // best-effort; ignore error for now
	}
	return access, refresh, nil
}
