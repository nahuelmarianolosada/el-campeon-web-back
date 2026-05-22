package guest

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/email"
	internalErr "github.com/nahuelmarianolosada/el-campeon-web/internal/services/errors"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type GuestService interface {
	VerifyEmailAndSendCode(emailAddr, clientIP string) error
	ConfirmEmailAndCreateSession(emailAddr, code, clientIP string) (*models.GuestSessionResponse, error)
	ValidateGuestToken(token, clientIP string) (*models.GuestSession, error)
	CreateAnonymousUser(email, firstName, lastName, phone string) (*uint, error)
}

type guestService struct {
	guestRepo    repositories.GuestRepository
	userRepo     repositories.UserRepository
	emailService email.EmailService
	config       *config.Config
}

func NewGuestService(
	guestRepo repositories.GuestRepository,
	userRepo repositories.UserRepository,
	emailService email.EmailService,
	cfg *config.Config,
) GuestService {
	return &guestService{
		guestRepo:    guestRepo,
		userRepo:     userRepo,
		emailService: emailService,
		config:       cfg,
	}
}

// VerifyEmailAndSendCode valida email y envía código de verificación
func (s *guestService) VerifyEmailAndSendCode(emailAddr, clientIP string) error {
	log.Printf("[guestService.VerifyEmailAndSendCode] INFO: Starting email verification - email=%s, ip=%s", emailAddr, clientIP)

	// Normalizar email
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))

	// Validar rate limiting por IP (max 3 intentos en 15 min)
	attempts, err := s.guestRepo.CountVerificationAttemptsByIP(clientIP, 15)
	if err != nil {
		log.Printf("[guestService.VerifyEmailAndSendCode] ERROR: Failed to count attempts - ip=%s: %v", clientIP, err)
		return fmt.Errorf("error checking rate limit: %w", err)
	}
	if attempts >= 3 {
		log.Printf("[guestService.VerifyEmailAndSendCode] WARNING: Rate limit exceeded - ip=%s, attempts=%d", clientIP, attempts)
		return fmt.Errorf("too many verification attempts from this IP. Try again in 15 minutes")
	}

	// Buscar o crear sesión guest
	session, err := s.guestRepo.FindGuestSessionByEmail(emailAddr)
	if err != nil {
		// Crear nueva sesión
		code := s.generateVerificationCode()
		codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[guestService.VerifyEmailAndSendCode] ERROR: Failed to hash code: %v", err)
			return fmt.Errorf("error generating verification code: %w", err)
		}

		now := time.Now()
		session = &models.GuestSession{
			Email:                    emailAddr,
			VerificationCodeHash:     string(codeHash),
			VerificationCodeSentAt:   &now,
			VerificationCodeAttempts: 0,
			IsVerified:               false,
			SessionIPAddress:         clientIP,
			AttemptsFromIP:           1,
			LastAttemptAt:            &now,
			ExpiresAt:                time.Now().AddDate(0, 0, 7),
		}

		if err := s.guestRepo.CreateGuestSession(session); err != nil {
			log.Printf("[guestService.VerifyEmailAndSendCode] ERROR: Failed to create session: %v", err)
			return fmt.Errorf("error creating guest session: %w", err)
		}

		// Enviar código por email
		if err := s.emailService.SendVerificationCode(emailAddr, code); err != nil {
			log.Printf("[guestService.VerifyEmailAndSendCode] ERROR: Failed to send email - email=%s: %v", emailAddr, err)
			return fmt.Errorf("error sending verification code: %w", err)
		}

		log.Printf("[guestService.VerifyEmailAndSendCode] INFO: Verification code sent - sessionID=%d, email=%s", session.ID, emailAddr)
		return nil
	}

	// Sesión existe, actualizar y reenviar código
	if session.IsVerified {
		log.Printf("[guestService.VerifyEmailAndSendCode] WARNING: Email already verified - email=%s", emailAddr)
		return internalErr.ErrEmailAlreadyVerified
	}

	// Generar nuevo código
	code := s.generateVerificationCode()
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[guestService.VerifyEmailAndSendCode] ERROR: Failed to hash code: %v", err)
		return fmt.Errorf("error generating verification code: %w", err)
	}

	now := time.Now()
	session.VerificationCodeHash = string(codeHash)
	session.VerificationCodeSentAt = &now
	session.VerificationCodeAttempts = 0
	session.SessionIPAddress = clientIP
	session.AttemptsFromIP = attempts + 1
	session.LastAttemptAt = &now

	if err := s.guestRepo.UpdateGuestSession(session); err != nil {
		log.Printf("[guestService.VerifyEmailAndSendCode] ERROR: Failed to update session: %v", err)
		return fmt.Errorf("error updating guest session: %w", err)
	}

	// Enviar código por email
	if err := s.emailService.SendVerificationCode(emailAddr, code); err != nil {
		log.Printf("[guestService.VerifyEmailAndSendCode] ERROR: Failed to send email - email=%s: %v", emailAddr, err)
		return fmt.Errorf("error sending verification code: %w", err)
	}

	log.Printf("[guestService.VerifyEmailAndSendCode] INFO: Verification code sent - sessionID=%d, email=%s", session.ID, emailAddr)
	return nil
}

// ConfirmEmailAndCreateSession valida código y crea sesión con token
func (s *guestService) ConfirmEmailAndCreateSession(emailAddr, code, clientIP string) (*models.GuestSessionResponse, error) {
	log.Printf("[guestService.ConfirmEmailAndCreateSession] INFO: Confirming email - email=%s, ip=%s", emailAddr, clientIP)

	// Normalizar email
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))

	// Buscar sesión
	session, err := s.guestRepo.FindGuestSessionByEmail(emailAddr)
	if err != nil {
		log.Printf("[guestService.ConfirmEmailAndCreateSession] ERROR: Session not found - email=%s", emailAddr)
		return nil, fmt.Errorf("email not found or not verified yet")
	}

	// Validar que no esté expirada
	if time.Now().After(session.ExpiresAt) {
		log.Printf("[guestService.ConfirmEmailAndCreateSession] WARNING: Session expired - email=%s", emailAddr)
		return nil, fmt.Errorf("verification session has expired")
	}

	// Validar código (max 3 intentos)
	if session.VerificationCodeAttempts >= 3 {
		log.Printf("[guestService.ConfirmEmailAndCreateSession] WARNING: Max verification attempts - email=%s", emailAddr)
		return nil, fmt.Errorf("too many verification attempts")
	}

	// Verificar código
	if err := bcrypt.CompareHashAndPassword([]byte(session.VerificationCodeHash), []byte(code)); err != nil {
		log.Printf("[guestService.ConfirmEmailAndCreateSession] WARNING: Invalid verification code - email=%s", emailAddr)
		session.VerificationCodeAttempts++
		if err := s.guestRepo.UpdateGuestSession(session); err != nil {
			log.Printf("[guestService.ConfirmEmailAndCreateSession] ERROR: Failed to update session: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("invalid verification code")
	}

	// Crear usuario anónimo
	userID, err := s.CreateAnonymousUser(emailAddr, "", "", "")
	if err != nil {
		log.Printf("[guestService.ConfirmEmailAndCreateSession] ERROR: Failed to create anonymous user - email=%s: %v", emailAddr, err)
		return nil, fmt.Errorf("error creating guest user: %w", err)
	}

	// Generar JWT tipo guest
	guestToken, err := utils.GenerateGuestToken(*userID, emailAddr, s.config)
	if err != nil {
		log.Printf("[guestService.ConfirmEmailAndCreateSession] ERROR: Failed to generate token - email=%s: %v", emailAddr, err)
		return nil, fmt.Errorf("error generating session token: %w", err)
	}

	// Hash token con IP
	tokenHash := s.hashTokenWithIP(guestToken, clientIP)

	// Actualizar sesión
	verifiedAt := time.Now()
	session.IsVerified = true
	session.VerifiedAt = &verifiedAt
	session.UserID = userID
	session.GuestTokenHash = tokenHash
	session.SessionIPAddress = clientIP
	session.VerificationCodeAttempts = 0

	if err := s.guestRepo.UpdateGuestSession(session); err != nil {
		log.Printf("[guestService.ConfirmEmailAndCreateSession] ERROR: Failed to update session: %v", err)
		return nil, fmt.Errorf("error updating session: %w", err)
	}

	log.Printf("[guestService.ConfirmEmailAndCreateSession] INFO: Email confirmed and session created - sessionID=%d, email=%s, userID=%d", session.ID, emailAddr, *userID)

	return &models.GuestSessionResponse{
		GuestToken: guestToken,
		Email:      emailAddr,
		ExpiresAt:  session.ExpiresAt,
	}, nil
}

// ValidateGuestToken valida un token guest
func (s *guestService) ValidateGuestToken(token, clientIP string) (*models.GuestSession, error) {
	log.Printf("[guestService.ValidateGuestToken] INFO: Validating guest token - ip=%s", clientIP)

	// Validar JWT
	claims, err := utils.ValidateGuestToken(token, s.config)
	if err != nil {
		log.Printf("[guestService.ValidateGuestToken] ERROR: Invalid token: %v", err)
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Buscar sesión por email
	session, err := s.guestRepo.FindGuestSessionByEmail(claims.Email)
	if err != nil {
		log.Printf("[guestService.ValidateGuestToken] ERROR: Session not found - email=%s", claims.Email)
		return nil, fmt.Errorf("session not found")
	}

	// Validar hash token + IP
	expectedHash := s.hashTokenWithIP(token, clientIP)
	if expectedHash != session.GuestTokenHash {
		log.Printf("[guestService.ValidateGuestToken] WARNING: Token hash mismatch - email=%s", claims.Email)
		return nil, fmt.Errorf("token validation failed: IP mismatch")
	}

	// Validar no expirado
	if time.Now().After(session.ExpiresAt) {
		log.Printf("[guestService.ValidateGuestToken] WARNING: Session expired - email=%s", claims.Email)
		return nil, fmt.Errorf("session has expired")
	}

	log.Printf("[guestService.ValidateGuestToken] INFO: Token valid - sessionID=%d, email=%s", session.ID, claims.Email)
	return session, nil
}

// CreateAnonymousUser crea un usuario anónimo temporal
func (s *guestService) CreateAnonymousUser(email, firstName, lastName, phone string) (*uint, error) {
	log.Printf("[guestService.CreateAnonymousUser] INFO: Creating anonymous user - email=%s", email)

	// Email normalizado
	email = strings.TrimSpace(strings.ToLower(email))

	userFound, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[guestService.CreateAnonymousUser] ERROR: Failed to find user - email=%s: %v", email, err)
			return nil, fmt.Errorf("error finding user: %w", err)
		}
	}

	if userFound != nil {
		log.Printf("[guestService.CreateAnonymousUser] WARNING: User already exists - email=%s", email)
		return &userFound.ID, nil
	}

	// Crear usuario con flag is_anonymous
	user := &models.User{
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Password:    "", // Sin contraseña
		Phone:       phone,
		IsAnonymous: true,
		IsActive:    true,
	}

	if err := s.userRepo.Create(user); err != nil {
		log.Printf("[guestService.CreateAnonymousUser] ERROR: Failed to create user: %v", err)
		return nil, fmt.Errorf("error creating anonymous user: %w", err)
	}

	log.Printf("[guestService.CreateAnonymousUser] INFO: Anonymous user created - userID=%d, email=%s", user.ID, email)
	return &user.ID, nil
}

// generateVerificationCode genera un código de 6 dígitos
func (s *guestService) generateVerificationCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", num.Int64())
	}
	return code
}

// hashTokenWithIP crea un hash SHA256 del token + IP
func (s *guestService) hashTokenWithIP(token, ip string) string {
	hash := sha256.Sum256([]byte(token + ip))
	return fmt.Sprintf("%x", hash)
}
