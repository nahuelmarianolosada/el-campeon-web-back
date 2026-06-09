package email

import (
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/stretchr/testify/assert"
)

// Si falta cualquiera de SMTPHost/User/Password, debe degradarse al noOpEmailService
// en lugar de fallar el arranque. Cubre las tres ramas del OR en NewEmailService.
func TestNewEmailService_MissingConfigReturnsNoOp(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.Config
	}{
		{"missing host", &config.Config{SMTPHost: "", SMTPUser: "u", SMTPPassword: "p"}},
		{"missing user", &config.Config{SMTPHost: "h", SMTPUser: "", SMTPPassword: "p"}},
		{"missing password", &config.Config{SMTPHost: "h", SMTPUser: "u", SMTPPassword: ""}},
		{"all empty", &config.Config{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := NewEmailService(tc.cfg)
			assert.NoError(t, err)
			_, isNoOp := svc.(*noOpEmailService)
			assert.True(t, isNoOp, "expected noOpEmailService when SMTP config incomplete")
		})
	}
}

func TestNewEmailService_AllConfigured_ReturnsGomailService(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:      "smtp.example.com",
		SMTPPort:      587,
		SMTPUser:      "user@example.com",
		SMTPPassword:  "secret",
		SMTPFromEmail: "from@example.com",
	}

	svc, err := NewEmailService(cfg)
	assert.NoError(t, err)
	_, isReal := svc.(*gomailService)
	assert.True(t, isReal, "expected real gomailService when SMTP fully configured")
}

func TestNoOpEmailService_SendVerificationCode_NoError(t *testing.T) {
	svc := &noOpEmailService{}
	assert.NoError(t, svc.SendVerificationCode("anyone@example.com", "123456"))
}

func TestNoOpEmailService_SendOrderConfirmation_NoError(t *testing.T) {
	svc := &noOpEmailService{}
	assert.NoError(t, svc.SendOrderConfirmation("anyone@example.com", "ORD-1", 100.50))
}
