package guest

import (
	"errors"
	"testing"
	"time"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MockGuestRepository struct {
	mock.Mock
}

func (m *MockGuestRepository) CreateGuestSession(session *models.GuestSession) error {
	args := m.Called(session)
	session.ID = 1
	return args.Error(0)
}

func (m *MockGuestRepository) FindGuestSessionByEmail(email string) (*models.GuestSession, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GuestSession), args.Error(1)
}

func (m *MockGuestRepository) UpdateGuestSession(session *models.GuestSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockGuestRepository) CountVerificationAttemptsByIP(ip string, window int) (int, error) {
	args := m.Called(ip, window)
	return args.Int(0), args.Error(1)
}

func (m *MockGuestRepository) DeleteExpiredSessions() (int64, error) {
	args := m.Called()
	return int64(args.Int(0)), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	user.ID = 1
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAll(limit, offset int) ([]models.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.User), args.Error(1)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendVerificationCode(email, code string) error {
	args := m.Called(email, code)
	return args.Error(0)
}

func (m *MockEmailService) SendOrderConfirmation(email, orderNumber string, total float64) error {
	args := m.Called(email, orderNumber, total)
	return args.Error(0)
}

func newTestService() (*MockGuestRepository, *MockUserRepository, *MockEmailService, GuestService) {
	repo := new(MockGuestRepository)
	userRepo := new(MockUserRepository)
	emailSvc := new(MockEmailService)
	cfg := &config.Config{JWTSecretKey: "test-secret", JWTExpiryHours: 24}
	svc := NewGuestService(repo, userRepo, emailSvc, cfg)
	return repo, userRepo, emailSvc, svc
}

func TestVerifyEmailAndSendCode_NewSession(t *testing.T) {
	repo, _, emailSvc, svc := newTestService()

	emailAddr := "test@example.com"
	clientIP := "127.0.0.1"

	repo.On("CountVerificationAttemptsByIP", clientIP, 15).Return(0, nil)
	repo.On("FindGuestSessionByEmail", emailAddr).Return(nil, assert.AnError)
	repo.On("CreateGuestSession", mock.Anything).Return(nil)
	emailSvc.On("SendVerificationCode", emailAddr, mock.Anything).Return(nil)

	err := svc.VerifyEmailAndSendCode(emailAddr, clientIP)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

// Re-verificación cuando ya existe sesión verificada: debe emitir un código nuevo,
// invalidar el token previo, resetear IsVerified/VerifiedAt y extender ExpiresAt.
// Regresión del bug donde el flujo se cortaba con ErrEmailAlreadyVerified.
func TestVerifyEmailAndSendCode_ExistingVerifiedSession_ReVerifies(t *testing.T) {
	repo, _, emailSvc, svc := newTestService()

	emailAddr := "verified@example.com"
	clientIP := "10.0.0.1"
	oldVerifiedAt := time.Now().Add(-48 * time.Hour)
	oldExpiresAt := time.Now().Add(24 * time.Hour)
	oldTokenHash := "old-token-hash-deadbeef"

	existing := &models.GuestSession{
		ID:              3,
		Email:           emailAddr,
		IsVerified:      true,
		VerifiedAt:      &oldVerifiedAt,
		GuestTokenHash:  oldTokenHash,
		ExpiresAt:       oldExpiresAt,
		VerificationCodeAttempts: 0,
	}

	repo.On("CountVerificationAttemptsByIP", clientIP, 15).Return(0, nil)
	repo.On("FindGuestSessionByEmail", emailAddr).Return(existing, nil)

	// Capturar el código generado para verificar que sea el mismo enviado por email.
	var sentCode string
	emailSvc.On("SendVerificationCode", emailAddr, mock.MatchedBy(func(code string) bool {
		sentCode = code
		return len(code) == 6
	})).Return(nil)

	repo.On("UpdateGuestSession", mock.MatchedBy(func(s *models.GuestSession) bool {
		return s.IsVerified == false &&
			s.VerifiedAt == nil &&
			s.GuestTokenHash == "" &&
			s.VerificationCodeHash != "" &&
			s.VerificationCodeAttempts == 0 &&
			s.ExpiresAt.After(oldExpiresAt)
	})).Return(nil)

	err := svc.VerifyEmailAndSendCode(emailAddr, clientIP)
	assert.NoError(t, err)
	assert.Len(t, sentCode, 6, "código enviado debe tener 6 dígitos")

	// El hash bcrypt persistido debe corresponder al código enviado.
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(existing.VerificationCodeHash), []byte(sentCode)))

	repo.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

func TestVerifyEmailAndSendCode_ExistingUnverifiedSession_Resends(t *testing.T) {
	repo, _, emailSvc, svc := newTestService()

	emailAddr := "pending@example.com"
	clientIP := "10.0.0.2"

	existing := &models.GuestSession{
		ID:         5,
		Email:      emailAddr,
		IsVerified: false,
		ExpiresAt:  time.Now().Add(72 * time.Hour),
	}

	repo.On("CountVerificationAttemptsByIP", clientIP, 15).Return(1, nil)
	repo.On("FindGuestSessionByEmail", emailAddr).Return(existing, nil)
	repo.On("UpdateGuestSession", mock.Anything).Return(nil)
	emailSvc.On("SendVerificationCode", emailAddr, mock.Anything).Return(nil)

	err := svc.VerifyEmailAndSendCode(emailAddr, clientIP)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

func TestVerifyEmailAndSendCode_RateLimitExceeded(t *testing.T) {
	repo, _, emailSvc, svc := newTestService()

	clientIP := "1.2.3.4"
	repo.On("CountVerificationAttemptsByIP", clientIP, 15).Return(3, nil)

	err := svc.VerifyEmailAndSendCode("rl@example.com", clientIP)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many verification attempts")

	repo.AssertExpectations(t)
	emailSvc.AssertNotCalled(t, "SendVerificationCode", mock.Anything, mock.Anything)
}

// El email debe normalizarse a lowercase y sin espacios antes de buscar/persistir.
func TestVerifyEmailAndSendCode_NormalizesEmail(t *testing.T) {
	repo, _, emailSvc, svc := newTestService()

	clientIP := "10.0.0.3"
	normalized := "mixedcase@example.com"

	repo.On("CountVerificationAttemptsByIP", clientIP, 15).Return(0, nil)
	repo.On("FindGuestSessionByEmail", normalized).Return(nil, assert.AnError)
	repo.On("CreateGuestSession", mock.MatchedBy(func(s *models.GuestSession) bool {
		return s.Email == normalized
	})).Return(nil)
	emailSvc.On("SendVerificationCode", normalized, mock.Anything).Return(nil)

	err := svc.VerifyEmailAndSendCode("  MixedCase@Example.COM  ", clientIP)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

func TestVerifyEmailAndSendCode_EmailServiceFailure_NewSession(t *testing.T) {
	repo, _, emailSvc, svc := newTestService()

	emailAddr := "smtpfail@example.com"
	clientIP := "10.0.0.4"

	repo.On("CountVerificationAttemptsByIP", clientIP, 15).Return(0, nil)
	repo.On("FindGuestSessionByEmail", emailAddr).Return(nil, assert.AnError)
	repo.On("CreateGuestSession", mock.Anything).Return(nil)
	emailSvc.On("SendVerificationCode", emailAddr, mock.Anything).Return(errors.New("smtp down"))

	err := svc.VerifyEmailAndSendCode(emailAddr, clientIP)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error sending verification code")
	repo.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

func TestVerifyEmailAndSendCode_RateLimitRepoFailure(t *testing.T) {
	repo, _, _, svc := newTestService()

	clientIP := "10.0.0.5"
	repo.On("CountVerificationAttemptsByIP", clientIP, 15).Return(0, errors.New("db down"))

	err := svc.VerifyEmailAndSendCode("any@example.com", clientIP)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
	repo.AssertExpectations(t)
}

func TestConfirmEmailAndCreateSession_SessionNotFound(t *testing.T) {
	repo, _, _, svc := newTestService()

	repo.On("FindGuestSessionByEmail", "missing@example.com").Return(nil, errors.New("not found"))

	resp, err := svc.ConfirmEmailAndCreateSession("missing@example.com", "123456", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
}

func TestConfirmEmailAndCreateSession_SessionExpired(t *testing.T) {
	repo, _, _, svc := newTestService()

	emailAddr := "expired@example.com"
	expired := &models.GuestSession{
		ID:        1,
		Email:     emailAddr,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	repo.On("FindGuestSessionByEmail", emailAddr).Return(expired, nil)

	resp, err := svc.ConfirmEmailAndCreateSession(emailAddr, "123456", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "expired")
}

func TestConfirmEmailAndCreateSession_TooManyAttempts(t *testing.T) {
	repo, _, _, svc := newTestService()

	emailAddr := "locked@example.com"
	session := &models.GuestSession{
		ID:                       1,
		Email:                    emailAddr,
		ExpiresAt:                time.Now().Add(time.Hour),
		VerificationCodeAttempts: 3,
	}
	repo.On("FindGuestSessionByEmail", emailAddr).Return(session, nil)

	resp, err := svc.ConfirmEmailAndCreateSession(emailAddr, "123456", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "too many verification attempts")
}

func TestConfirmEmailAndCreateSession_InvalidCode_IncrementsAttempts(t *testing.T) {
	repo, _, _, svc := newTestService()

	emailAddr := "wrong@example.com"
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	session := &models.GuestSession{
		ID:                       1,
		Email:                    emailAddr,
		ExpiresAt:                time.Now().Add(time.Hour),
		VerificationCodeHash:     string(hash),
		VerificationCodeAttempts: 0,
	}
	repo.On("FindGuestSessionByEmail", emailAddr).Return(session, nil)
	repo.On("UpdateGuestSession", mock.MatchedBy(func(s *models.GuestSession) bool {
		return s.VerificationCodeAttempts == 1
	})).Return(nil)

	resp, err := svc.ConfirmEmailAndCreateSession(emailAddr, "999999", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid verification code")
	repo.AssertExpectations(t)
}

func TestConfirmEmailAndCreateSession_Success(t *testing.T) {
	repo, userRepo, _, svc := newTestService()

	emailAddr := "ok@example.com"
	code := "654321"
	hash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	session := &models.GuestSession{
		ID:                   1,
		Email:                emailAddr,
		ExpiresAt:            time.Now().Add(time.Hour),
		VerificationCodeHash: string(hash),
	}

	repo.On("FindGuestSessionByEmail", emailAddr).Return(session, nil)
	userRepo.On("FindByEmail", emailAddr).Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("Create", mock.MatchedBy(func(u *models.User) bool {
		return u.Email == emailAddr && u.IsAnonymous && u.IsActive
	})).Return(nil)
	repo.On("UpdateGuestSession", mock.MatchedBy(func(s *models.GuestSession) bool {
		return s.IsVerified && s.VerifiedAt != nil && s.GuestTokenHash != "" && s.UserID != nil
	})).Return(nil)

	resp, err := svc.ConfirmEmailAndCreateSession(emailAddr, code, "127.0.0.1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, emailAddr, resp.Email)
	assert.NotEmpty(t, resp.GuestToken)
	repo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreateAnonymousUser_ExistingUser_ReturnsExistingID(t *testing.T) {
	_, userRepo, _, svc := newTestService()

	existing := &models.User{Email: "exists@example.com"}
	existing.ID = 42
	userRepo.On("FindByEmail", "exists@example.com").Return(existing, nil)

	id, err := svc.CreateAnonymousUser("exists@example.com", "", "", "")
	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, uint(42), *id)
	userRepo.AssertExpectations(t)
	userRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestCreateAnonymousUser_NewUser_CreatesIt(t *testing.T) {
	_, userRepo, _, svc := newTestService()

	userRepo.On("FindByEmail", "new@example.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("Create", mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "new@example.com" && u.IsAnonymous && u.IsActive && u.Password == ""
	})).Return(nil)

	id, err := svc.CreateAnonymousUser("new@example.com", "", "", "")
	assert.NoError(t, err)
	assert.NotNil(t, id)
	userRepo.AssertExpectations(t)
}

// Errores de DB distintos a "not found" deben propagarse, no silenciarse como nuevo usuario.
func TestCreateAnonymousUser_RepoErrorPropagates(t *testing.T) {
	_, userRepo, _, svc := newTestService()

	userRepo.On("FindByEmail", "boom@example.com").Return(nil, errors.New("db connection lost"))

	id, err := svc.CreateAnonymousUser("boom@example.com", "", "", "")
	assert.Error(t, err)
	assert.Nil(t, id)
	userRepo.AssertExpectations(t)
}

func TestValidateGuestToken_InvalidToken(t *testing.T) {
	_, _, _, svc := newTestService()

	session, err := svc.ValidateGuestToken("not-a-jwt", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, session)
}
