package services

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/GyBJluHv2/sensory-navigator/backend/internal/auth"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/config"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/email"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/models"
	"gorm.io/gorm"
)

// MaxVerificationAttempts ограничивает число попыток ввода кода для одной записи.
// После исчерпания запись помечается использованной и пользователю требуется
// запросить новый код.
const MaxVerificationAttempts = 5

var (
	ErrInvalidEmail        = errors.New("введите корректный email")
	ErrCodeNotFound        = errors.New("активный код подтверждения не найден")
	ErrCodeExpired         = errors.New("код подтверждения истёк, запросите новый")
	ErrCodeAttemptsLimit   = errors.New("слишком много попыток, запросите новый код")
	ErrCodeMismatch        = errors.New("неверный код подтверждения")
	ErrAlreadyConfirmed    = errors.New("учётная запись уже подтверждена")
)

// VerificationService — сервис подтверждения email кодом из 6 цифр.
type VerificationService struct {
	db     *gorm.DB
	cfg    *config.Config
	users  *UserService
	mailer email.Sender
}

func NewVerificationService(db *gorm.DB, cfg *config.Config, users *UserService, mailer email.Sender) *VerificationService {
	return &VerificationService{db: db, cfg: cfg, users: users, mailer: mailer}
}

// IsEmailFormatValid проверяет email по умеренно строгому регулярному
// выражению: одна @, по крайней мере одна точка в домене, без пробелов.
func IsEmailFormatValid(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" || len(email) > 255 {
		return false
	}
	at := strings.LastIndex(email, "@")
	if at < 1 || at >= len(email)-3 {
		return false
	}
	if strings.Contains(email, " ") {
		return false
	}
	domain := email[at+1:]
	if !strings.Contains(domain, ".") {
		return false
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	return true
}

// RequestRegisterInput — данные для шага 1 регистрации (отправка кода).
type RequestRegisterInput struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
}

// Request создаёт (или обновляет) запись VerificationCode и отправляет код
// на указанный email. Учётная запись пользователя на этом шаге ещё не
// создаётся: пароль временно хранится в bcrypt-хэше внутри VerificationCode
// и переезжает в users только при успешном подтверждении кода.
func (s *VerificationService) Request(in RequestRegisterInput) (time.Time, error) {
	if !IsEmailFormatValid(in.Email) {
		return time.Time{}, ErrInvalidEmail
	}
	if len(in.Password) < 6 {
		return time.Time{}, errors.New("пароль должен быть не короче 6 символов")
	}
	if len(in.Username) < 3 {
		return time.Time{}, errors.New("имя пользователя должно быть не короче 3 символов")
	}

	// Проверка, что email/username ещё не заняты подтверждённым пользователем.
	var n int64
	s.db.Model(&models.User{}).
		Where("email = ? AND email_verified = TRUE", in.Email).
		Count(&n)
	if n > 0 {
		return time.Time{}, ErrEmailTaken
	}
	s.db.Model(&models.User{}).
		Where("username = ? AND email_verified = TRUE", in.Username).
		Count(&n)
	if n > 0 {
		return time.Time{}, ErrUsernameTaken
	}

	code, err := generateNumericCode(6)
	if err != nil {
		return time.Time{}, err
	}

	codeHash, err := auth.GeneratePasswordHash(code, 4)
	if err != nil {
		return time.Time{}, err
	}
	pwdHash, err := auth.GeneratePasswordHash(in.Password, s.cfg.BCryptCost)
	if err != nil {
		return time.Time{}, err
	}

	expiresAt := time.Now().Add(s.cfg.VerificationTTL)

	// Деактивируем все предыдущие активные коды на этот email.
	now := time.Now()
	s.db.Model(&models.VerificationCode{}).
		Where("email = ? AND used_at IS NULL AND expires_at > ?", in.Email, now).
		Update("used_at", now)

	rec := &models.VerificationCode{
		Email:        in.Email,
		CodeHash:     codeHash,
		PasswordHash: pwdHash,
		Username:     in.Username,
		DisplayName:  in.DisplayName,
		ExpiresAt:    expiresAt,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return time.Time{}, err
	}

	if err := s.mailer.SendVerificationCode(in.Email, code); err != nil {
		// Если отправка не удалась — сразу инвалидируем код, чтобы не висел в БД.
		s.db.Model(rec).Update("used_at", time.Now())
		return time.Time{}, err
	}
	return expiresAt, nil
}

// Confirm проверяет код и при успехе создаёт учётную запись с email_verified=TRUE.
// Возвращает свежесозданного User для последующей выдачи JWT.
func (s *VerificationService) Confirm(emailAddr, code string) (*models.User, error) {
	emailAddr = strings.TrimSpace(emailAddr)
	if !IsEmailFormatValid(emailAddr) {
		return nil, ErrInvalidEmail
	}

	var rec models.VerificationCode
	err := s.db.Where("email = ? AND used_at IS NULL", emailAddr).
		Order("created_at DESC").
		First(&rec).Error
	if err != nil {
		return nil, ErrCodeNotFound
	}
	if time.Now().After(rec.ExpiresAt) {
		return nil, ErrCodeExpired
	}
	if rec.Attempts >= MaxVerificationAttempts {
		now := time.Now()
		s.db.Model(&rec).Update("used_at", now)
		return nil, ErrCodeAttemptsLimit
	}

	if !auth.CheckPassword(code, rec.CodeHash) {
		s.db.Model(&rec).UpdateColumn("attempts", gorm.Expr("attempts + 1"))
		return nil, ErrCodeMismatch
	}

	// Дополнительно убеждаемся, что email/username не успели "захватить" другие.
	var n int64
	s.db.Model(&models.User{}).
		Where("email = ? AND email_verified = TRUE", rec.Email).
		Count(&n)
	if n > 0 {
		return nil, ErrEmailTaken
	}
	s.db.Model(&models.User{}).
		Where("username = ? AND email_verified = TRUE", rec.Username).
		Count(&n)
	if n > 0 {
		return nil, ErrUsernameTaken
	}

	user := &models.User{
		Email:         rec.Email,
		Username:      rec.Username,
		PasswordHash:  rec.PasswordHash,
		DisplayName:   rec.DisplayName,
		EmailVerified: true,
		NoisePref:     3,
		LightPref:     3,
		CrowdPref:     3,
	}
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	s.db.Model(&rec).Update("used_at", now)
	return user, nil
}

// Resend генерирует новый код и отправляет его повторно.
// Используется, когда первое письмо не дошло.
func (s *VerificationService) Resend(emailAddr string) (time.Time, error) {
	emailAddr = strings.TrimSpace(emailAddr)
	if !IsEmailFormatValid(emailAddr) {
		return time.Time{}, ErrInvalidEmail
	}

	var rec models.VerificationCode
	err := s.db.Where("email = ? AND used_at IS NULL", emailAddr).
		Order("created_at DESC").
		First(&rec).Error
	if err != nil {
		return time.Time{}, ErrCodeNotFound
	}

	code, err := generateNumericCode(6)
	if err != nil {
		return time.Time{}, err
	}
	codeHash, err := auth.GeneratePasswordHash(code, 4)
	if err != nil {
		return time.Time{}, err
	}
	expiresAt := time.Now().Add(s.cfg.VerificationTTL)

	rec.CodeHash = codeHash
	rec.ExpiresAt = expiresAt
	rec.Attempts = 0
	if err := s.db.Save(&rec).Error; err != nil {
		return time.Time{}, err
	}
	if err := s.mailer.SendVerificationCode(emailAddr, code); err != nil {
		return time.Time{}, err
	}
	return expiresAt, nil
}

// generateNumericCode возвращает строку из n десятичных цифр со случайными
// значениями, полученными через crypto/rand. Подходит для одноразовых
// кодов подтверждения email.
func generateNumericCode(n int) (string, error) {
	const digits = "0123456789"
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		out[i] = digits[idx.Int64()]
	}
	return string(out), nil
}