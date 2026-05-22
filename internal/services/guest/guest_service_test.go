package guest

import (
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestVerifyEmailAndSendCode_NewSession(t *testing.T) {
	repo := new(MockGuestRepository)
	userRepo := new(MockUserRepository)
	emailSvc := new(MockEmailService)
	cfg := &config.Config{JWTSecretKey: "test-secret"}
	svc := NewGuestService(repo, userRepo, emailSvc, cfg)

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
