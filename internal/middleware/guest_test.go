package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestGuestAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	svc := new(mockGuestService)
	r := gin.New()
	r.Use(GuestAuthMiddleware(svc))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing X-Guest-Token")
}

func TestGuestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	svc := new(mockGuestService)
	svc.On("ValidateGuestToken", "bad-token", mock.Anything).Return(nil, errors.New("invalid"))

	r := gin.New()
	r.Use(GuestAuthMiddleware(svc))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Guest-Token", "bad-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired guest token")
	svc.AssertExpectations(t)
}

func TestGuestAuthMiddleware_ValidToken_PopulatesContext(t *testing.T) {
	svc := new(mockGuestService)
	uid := uint(7)
	session := &models.GuestSession{
		Email:  "g@example.com",
		UserID: &uid,
	}
	session.ID = 11
	svc.On("ValidateGuestToken", "good-token", mock.Anything).Return(session, nil)

	r := gin.New()
	r.Use(GuestAuthMiddleware(svc))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"guest_session_id": c.GetUint("guest_session_id"),
			"guest_email":      c.GetString("guest_email"),
		})
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Guest-Token", "good-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"guest_email":"g@example.com"`)
	assert.Contains(t, w.Body.String(), `"guest_session_id":11`)
	svc.AssertExpectations(t)
}

func TestOptionalAuthWithGuestMiddleware_NoHeaders_Continues(t *testing.T) {
	svc := new(mockGuestService)
	r := gin.New()
	r.Use(OptionalAuthWithGuestMiddleware(nil, svc))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuthWithGuestMiddleware_ValidGuestToken_SetsContext(t *testing.T) {
	svc := new(mockGuestService)
	uid := uint(3)
	session := &models.GuestSession{Email: "g@example.com", UserID: &uid}
	session.ID = 9
	svc.On("ValidateGuestToken", "g-token", mock.Anything).Return(session, nil)

	r := gin.New()
	r.Use(OptionalAuthWithGuestMiddleware(nil, svc))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"is_guest":    c.GetBool("is_guest"),
			"guest_email": c.GetString("guest_email"),
		})
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Guest-Token", "g-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_guest":true`)
	svc.AssertExpectations(t)
}

func TestOptionalAuthWithGuestMiddleware_InvalidGuestToken_FallsThrough(t *testing.T) {
	svc := new(mockGuestService)
	svc.On("ValidateGuestToken", "bad", mock.Anything).Return(nil, errors.New("invalid"))

	r := gin.New()
	r.Use(OptionalAuthWithGuestMiddleware(nil, svc))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"is_guest": c.GetBool("is_guest")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Guest-Token", "bad")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_guest":false`)
	svc.AssertExpectations(t)
}
