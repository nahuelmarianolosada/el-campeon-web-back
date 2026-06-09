package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockGuestService struct {
	mock.Mock
}

func (m *mockGuestService) VerifyEmailAndSendCode(email, ip string) error {
	args := m.Called(email, ip)
	return args.Error(0)
}

func (m *mockGuestService) ConfirmEmailAndCreateSession(email, code, ip string) (*models.GuestSessionResponse, error) {
	args := m.Called(email, code, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GuestSessionResponse), args.Error(1)
}

func (m *mockGuestService) ValidateGuestToken(token, ip string) (*models.GuestSession, error) {
	args := m.Called(token, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GuestSession), args.Error(1)
}

func (m *mockGuestService) CreateAnonymousUser(email, fn, ln, phone string) (*uint, error) {
	args := m.Called(email, fn, ln, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	id := args.Get(0).(uint)
	return &id, args.Error(1)
}

func newGuestRouter(svc *mockGuestService) *gin.Engine {
	h := NewGuestHandler(svc)
	r := gin.New()
	r.POST("/verify-email", h.VerifyEmail)
	r.POST("/confirm-email", h.ConfirmEmail)
	return r
}

func doJSON(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestVerifyEmail_Success(t *testing.T) {
	svc := new(mockGuestService)
	svc.On("VerifyEmailAndSendCode", "u@example.com", mock.Anything).Return(nil)

	w := doJSON(newGuestRouter(svc), "POST", "/verify-email", `{"email":"u@example.com"}`)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Verification code sent")
	assert.Contains(t, w.Body.String(), `"expires_in_seconds":600`)
	svc.AssertExpectations(t)
}

func TestVerifyEmail_InvalidJSON_Returns400(t *testing.T) {
	svc := new(mockGuestService)
	w := doJSON(newGuestRouter(svc), "POST", "/verify-email", `{not json`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "VerifyEmailAndSendCode", mock.Anything, mock.Anything)
}

func TestVerifyEmail_BadEmailFormat_Returns400(t *testing.T) {
	svc := new(mockGuestService)
	w := doJSON(newGuestRouter(svc), "POST", "/verify-email", `{"email":"not-an-email"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyEmail_RateLimit_Returns429(t *testing.T) {
	svc := new(mockGuestService)
	svc.On("VerifyEmailAndSendCode", "u@example.com", mock.Anything).Return(
		errors.New("too many verification attempts from this IP. Try again in 15 minutes"),
	)

	w := doJSON(newGuestRouter(svc), "POST", "/verify-email", `{"email":"u@example.com"}`)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	svc.AssertExpectations(t)
}

func TestVerifyEmail_GenericError_Returns500(t *testing.T) {
	svc := new(mockGuestService)
	svc.On("VerifyEmailAndSendCode", "u@example.com", mock.Anything).Return(errors.New("smtp boom"))

	w := doJSON(newGuestRouter(svc), "POST", "/verify-email", `{"email":"u@example.com"}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "smtp boom")
}

func TestConfirmEmail_Success(t *testing.T) {
	svc := new(mockGuestService)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	svc.On("ConfirmEmailAndCreateSession", "u@example.com", "123456", mock.Anything).Return(
		&models.GuestSessionResponse{
			GuestToken: "jwt-token",
			Email:      "u@example.com",
			ExpiresAt:  expiresAt,
		}, nil,
	)

	w := doJSON(newGuestRouter(svc), "POST", "/confirm-email",
		`{"email":"u@example.com","verification_code":"123456"}`)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"guest_token":"jwt-token"`)
	assert.Contains(t, w.Body.String(), `"email":"u@example.com"`)
	svc.AssertExpectations(t)
}

func TestConfirmEmail_InvalidJSON_Returns400(t *testing.T) {
	svc := new(mockGuestService)
	w := doJSON(newGuestRouter(svc), "POST", "/confirm-email", `not json`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfirmEmail_ShortCode_Returns400(t *testing.T) {
	svc := new(mockGuestService)
	// El binding requiere len=6 — un código de 4 caracteres debe rechazarse antes del service.
	w := doJSON(newGuestRouter(svc), "POST", "/confirm-email",
		`{"email":"u@example.com","verification_code":"1234"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "ConfirmEmailAndCreateSession", mock.Anything, mock.Anything, mock.Anything)
}

func TestConfirmEmail_ServiceError_Returns500(t *testing.T) {
	svc := new(mockGuestService)
	svc.On("ConfirmEmailAndCreateSession", "u@example.com", "123456", mock.Anything).Return(
		nil, errors.New("invalid verification code"),
	)

	w := doJSON(newGuestRouter(svc), "POST", "/confirm-email",
		`{"email":"u@example.com","verification_code":"123456"}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "invalid verification code")
}
