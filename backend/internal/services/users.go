package services

import (
	"errors"

	"github.com/atakhanov/sensory-navigator/backend/internal/auth"
	"github.com/atakhanov/sensory-navigator/backend/internal/config"
	"github.com/atakhanov/sensory-navigator/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken    = errors.New("пользователь с таким email уже зарегистрирован")
	ErrUsernameTaken = errors.New("имя пользователя уже занято")
	ErrBadCreds      = errors.New("неверный логин или пароль")
)

type UserService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewUserService(db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{db: db, cfg: cfg}
}

type RegisterInput struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
}

func (s *UserService) Register(in RegisterInput) (*models.User, error) {
	var n int64
	s.db.Model(&models.User{}).Where("email = ?", in.Email).Count(&n)
	if n > 0 {
		return nil, ErrEmailTaken
	}
	s.db.Model(&models.User{}).Where("username = ?", in.Username).Count(&n)
	if n > 0 {
		return nil, ErrUsernameTaken
	}

	hash, err := auth.GeneratePasswordHash(in.Password, s.cfg.BCryptCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: hash,
		DisplayName:  in.DisplayName,
		NoisePref:    3,
		LightPref:    3,
		CrowdPref:    3,
	}
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil || !auth.CheckPassword(password, user.PasswordHash) {
		return nil, ErrBadCreds
	}
	return &user, nil
}

func (s *UserService) GetByID(id uint64) (*models.User, error) {
	var u models.User
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

type UpdateProfileInput struct {
	DisplayName string
	AvatarURL   string
	NoisePref   int
	LightPref   int
	CrowdPref   int
}

func (s *UserService) UpdateProfile(userID uint64, in UpdateProfileInput) (*models.User, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}
	user.DisplayName = in.DisplayName
	user.AvatarURL = in.AvatarURL
	if in.NoisePref > 0 {
		user.NoisePref = in.NoisePref
	}
	if in.LightPref > 0 {
		user.LightPref = in.LightPref
	}
	if in.CrowdPref > 0 {
		user.CrowdPref = in.CrowdPref
	}
	if err := s.db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) ChangePassword(userID uint64, oldPw, newPw string) error {
	user, err := s.GetByID(userID)
	if err != nil {
		return err
	}
	if !auth.CheckPassword(oldPw, user.PasswordHash) {
		return ErrBadCreds
	}
	hash, err := auth.GeneratePasswordHash(newPw, s.cfg.BCryptCost)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.db.Save(user).Error
}