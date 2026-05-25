package email

import (
	"fmt"
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendVerificationCode(toEmail, code string) error
	SendOrderConfirmation(toEmail, orderNumber string, total float64) error
}

type gomailService struct {
	host      string
	port      int
	user      string
	password  string
	fromEmail string
	dialer    *gomail.Dialer
}

func NewEmailService(cfg *config.Config) (EmailService, error) {
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPassword == "" {
		log.Printf("[gomailService.NewEmailService] WARNING: SMTP not configured, using no-op service")
		return &noOpEmailService{}, nil
	}

	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)

	return &gomailService{
		host:      cfg.SMTPHost,
		port:      cfg.SMTPPort,
		user:      cfg.SMTPUser,
		password:  cfg.SMTPPassword,
		fromEmail: cfg.SMTPFromEmail,
		dialer:    dialer,
	}, nil
}

func (s *gomailService) SendVerificationCode(toEmail, code string) error {
	log.Printf("[gomailService.SendVerificationCode] INFO: Sending verification code - email=%s", toEmail)

	subject := "Tu código de verificación - El Campeón"
	text := fmt.Sprintf("Tu código de verificación es: %s\n\nEste código es válido por 10 minutos.", code)
	html := fmt.Sprintf(`
		<html>
			<body>
				<h2>Tu código de verificación</h2>
				<p>Tu código es: <strong>%s</strong></p>
				<p>Este código es válido por 10 minutos.</p>
				<p>Si no solicitaste este código, ignora este email.</p>
			</body>
		</html>
	`, code)

	m := gomail.NewMessage()
	m.SetHeader("From", s.fromEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", text)
	m.AddAlternative("text/html", html)

	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("[gomailService.SendVerificationCode] ERROR: Failed to send email - email=%s: %v", toEmail, err)
		return fmt.Errorf("failed to send verification code: %w", err)
	}

	log.Printf("[gomailService.SendVerificationCode] INFO: Verification code sent - email=%s", toEmail)
	return nil
}

func (s *gomailService) SendOrderConfirmation(toEmail, orderNumber string, total float64) error {
	log.Printf("[gomailService.SendOrderConfirmation] INFO: Sending order confirmation - email=%s, orderNumber=%s", toEmail, orderNumber)

	subject := fmt.Sprintf("Confirmación de Orden #%s - El Campeón", orderNumber)
	text := fmt.Sprintf("Tu orden #%s ha sido confirmada.\nTotal: $%.2f\n\nGracias por tu compra.", orderNumber, total)
	html := fmt.Sprintf(`
		<html>
			<body>
				<h2>Confirmación de Orden</h2>
				<p>Número de orden: <strong>#%s</strong></p>
				<p>Total: <strong>$%.2f</strong></p>
				<p>Gracias por tu compra en El Campeón.</p>
			</body>
		</html>
	`, orderNumber, total)

	m := gomail.NewMessage()
	m.SetHeader("From", s.fromEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", text)
	m.AddAlternative("text/html", html)

	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("[gomailService.SendOrderConfirmation] ERROR: Failed to send email - email=%s: %v", toEmail, err)
		return fmt.Errorf("failed to send order confirmation: %w", err)
	}

	log.Printf("[gomailService.SendOrderConfirmation] INFO: Order confirmation sent - email=%s", toEmail)
	return nil
}

// noOpEmailService es usado cuando el servicio de email no está configurado
type noOpEmailService struct{}

func (s *noOpEmailService) SendVerificationCode(toEmail, code string) error {
	log.Printf("[noOpEmailService.SendVerificationCode] WARNING: Email not configured, skipping email - email=%s, code=%s", toEmail, code)
	return nil
}

func (s *noOpEmailService) SendOrderConfirmation(toEmail, orderNumber string, total float64) error {
	log.Printf("[noOpEmailService.SendOrderConfirmation] WARNING: Email not configured, skipping email - email=%s, order=%s", toEmail, orderNumber)
	return nil
}
