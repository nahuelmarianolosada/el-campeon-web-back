package email

import (
	"fmt"
	"log"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
)

type EmailService interface {
	SendVerificationCode(toEmail, code string) error
	SendOrderConfirmation(toEmail, orderNumber string, total float64) error
}

type mailgunService struct {
	domain    string
	apiKey    string
	fromEmail string
	mg        *mailgun.MailgunImpl
}

func NewMailgunService(cfg *config.Config) (EmailService, error) {
	if cfg.MailgunDomain == "" || cfg.MailgunAPIKey == "" {
		log.Printf("[mailgunService.NewMailgunService] WARNING: Mailgun not configured, using no-op service")
		return &noOpService{}, nil
	}

	mg := mailgun.NewMailgun(cfg.MailgunDomain, cfg.MailgunAPIKey)

	return &mailgunService{
		domain:    cfg.MailgunDomain,
		apiKey:    cfg.MailgunAPIKey,
		fromEmail: cfg.MailgunFromEmail,
		mg:        mg,
	}, nil
}

func (s *mailgunService) SendVerificationCode(toEmail, code string) error {
	log.Printf("[mailgunService.SendVerificationCode] INFO: Sending verification code - email=%s", toEmail)

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

	message := s.mg.NewMessage(s.fromEmail, subject, text, toEmail)
	message.SetHtml(html)

	ctx, cancel := contextWithTimeout()
	defer cancel()

	_, _, err := s.mg.Send(ctx, message)
	if err != nil {
		log.Printf("[mailgunService.SendVerificationCode] ERROR: Failed to send email - email=%s: %v", toEmail, err)
		return fmt.Errorf("failed to send verification code: %w", err)
	}

	log.Printf("[mailgunService.SendVerificationCode] INFO: Verification code sent - email=%s", toEmail)
	return nil
}

func (s *mailgunService) SendOrderConfirmation(toEmail, orderNumber string, total float64) error {
	log.Printf("[mailgunService.SendOrderConfirmation] INFO: Sending order confirmation - email=%s, orderNumber=%s", toEmail, orderNumber)

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

	message := s.mg.NewMessage(s.fromEmail, subject, text, toEmail)
	message.SetHtml(html)

	ctx, cancel := contextWithTimeout()
	defer cancel()

	_, _, err := s.mg.Send(ctx, message)
	if err != nil {
		log.Printf("[mailgunService.SendOrderConfirmation] ERROR: Failed to send email - email=%s: %v", toEmail, err)
		return fmt.Errorf("failed to send order confirmation: %w", err)
	}

	log.Printf("[mailgunService.SendOrderConfirmation] INFO: Order confirmation sent - email=%s", toEmail)
	return nil
}

// noOpService es usado cuando Mailgun no está configurado
type noOpService struct{}

func (s *noOpService) SendVerificationCode(toEmail, code string) error {
	log.Printf("[noOpService.SendVerificationCode] WARNING: Mailgun not configured, skipping email - email=%s, code=%s", toEmail, code)
	return nil
}

func (s *noOpService) SendOrderConfirmation(toEmail, orderNumber string, total float64) error {
	log.Printf("[noOpService.SendOrderConfirmation] WARNING: Mailgun not configured, skipping email - email=%s, order=%s", toEmail, orderNumber)
	return nil
}

