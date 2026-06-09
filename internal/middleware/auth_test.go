package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newAuthCfg() *config.Config {
	return &config.Config{
		JWTSecretKey:     "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTExpiryHours:   1,
	}
}

func runRequest(handler gin.HandlerFunc, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	r := gin.New()
	r.Use(handler)
	r.Handle(method, path, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetUint("user_id"),
			"email":   c.GetString("email"),
			"role":    c.GetString("role"),
		})
	})

	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	w := runRequest(AuthMiddleware(newAuthCfg()), "GET", "/", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestAuthMiddleware_InvalidFormat_Returns401(t *testing.T) {
	cases := []string{
		"not-a-bearer-token",
		"Token abc",
		"Bearer",
		"Bearer one two",
	}
	for _, h := range cases {
		t.Run(h, func(t *testing.T) {
			w := runRequest(AuthMiddleware(newAuthCfg()), "GET", "/", map[string]string{"Authorization": h})
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "invalid authorization header format")
		})
	}
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	w := runRequest(AuthMiddleware(newAuthCfg()), "GET", "/", map[string]string{
		"Authorization": "Bearer not.a.jwt",
	})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	cfg := newAuthCfg()
	refresh, err := utils.GenerateRefreshToken(1, "u@example.com", "USER", cfg)
	assert.NoError(t, err)

	w := runRequest(AuthMiddleware(cfg), "GET", "/", map[string]string{
		"Authorization": "Bearer " + refresh,
	})
	// ValidateAccessToken rechaza refresh tokens.
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken_PopulatesContext(t *testing.T) {
	cfg := newAuthCfg()
	token, err := utils.GenerateAccessToken(42, "u@example.com", "USER", cfg)
	assert.NoError(t, err)

	w := runRequest(AuthMiddleware(cfg), "GET", "/", map[string]string{
		"Authorization": "Bearer " + token,
	})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"email":"u@example.com"`)
	assert.Contains(t, w.Body.String(), `"role":"USER"`)
}

func TestAdminMiddleware_NoRole_Returns403(t *testing.T) {
	r := gin.New()
	r.Use(AdminMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "role not found")
}

func TestAdminMiddleware_NonAdmin_Returns403(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", "USER"); c.Next() }, AdminMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
}

func TestAdminMiddleware_Admin_PassesThrough(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", "ADMIN"); c.Next() }, AdminMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuthMiddleware_NoHeader_Continues(t *testing.T) {
	w := runRequest(OptionalAuthMiddleware(newAuthCfg()), "GET", "/", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":0`)
}

func TestOptionalAuthMiddleware_InvalidHeader_ContinuesSilently(t *testing.T) {
	w := runRequest(OptionalAuthMiddleware(newAuthCfg()), "GET", "/", map[string]string{
		"Authorization": "Bearer garbage",
	})
	// El middleware opcional no debe abortar aunque el token sea inválido.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":0`)
}

func TestOptionalAuthMiddleware_ValidToken_PopulatesContext(t *testing.T) {
	cfg := newAuthCfg()
	token, err := utils.GenerateAccessToken(99, "opt@example.com", "USER", cfg)
	assert.NoError(t, err)

	w := runRequest(OptionalAuthMiddleware(cfg), "GET", "/", map[string]string{
		"Authorization": "Bearer " + token,
	})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"email":"opt@example.com"`)
}
