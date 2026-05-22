package repositories

import (
	"log"
	"time"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type GuestRepository interface {
	CreateGuestSession(session *models.GuestSession) error
	FindGuestSessionByEmail(email string) (*models.GuestSession, error)
	UpdateGuestSession(session *models.GuestSession) error
	CountVerificationAttemptsByIP(ip string, minutesWindow int) (int, error)
	DeleteExpiredSessions() (int64, error)
}

type guestRepository struct {
	db *gorm.DB
}

func NewGuestRepository(db *gorm.DB) GuestRepository {
	return &guestRepository{db: db}
}

func (r *guestRepository) CreateGuestSession(session *models.GuestSession) error {
	log.Printf("[guestRepository.CreateGuestSession] INFO: Creating guest session - email=%s", session.Email)
	if err := r.db.Create(session).Error; err != nil {
		log.Printf("[guestRepository.CreateGuestSession] ERROR: Failed to create session - email=%s: %v", session.Email, err)
		return err
	}
	log.Printf("[guestRepository.CreateGuestSession] INFO: Session created - sessionID=%d, email=%s", session.ID, session.Email)
	return nil
}

func (r *guestRepository) FindGuestSessionByEmail(email string) (*models.GuestSession, error) {
	log.Printf("[guestRepository.FindGuestSessionByEmail] INFO: Finding session - email=%s", email)
	var session models.GuestSession
	if err := r.db.Where("email = ?", email).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[guestRepository.FindGuestSessionByEmail] WARNING: Session not found - email=%s", email)
		} else {
			log.Printf("[guestRepository.FindGuestSessionByEmail] ERROR: Database error - email=%s: %v", email, err)
		}
		return nil, err
	}
	log.Printf("[guestRepository.FindGuestSessionByEmail] INFO: Session found - sessionID=%d, isVerified=%v", session.ID, session.IsVerified)
	return &session, nil
}

func (r *guestRepository) UpdateGuestSession(session *models.GuestSession) error {
	log.Printf("[guestRepository.UpdateGuestSession] INFO: Updating session - sessionID=%d, email=%s", session.ID, session.Email)
	if err := r.db.Save(session).Error; err != nil {
		log.Printf("[guestRepository.UpdateGuestSession] ERROR: Failed to update session - sessionID=%d: %v", session.ID, err)
		return err
	}
	log.Printf("[guestRepository.UpdateGuestSession] INFO: Session updated successfully - sessionID=%d", session.ID)
	return nil
}

func (r *guestRepository) CountVerificationAttemptsByIP(ip string, minutesWindow int) (int, error) {
	log.Printf("[guestRepository.CountVerificationAttemptsByIP] INFO: Counting attempts - ip=%s, window=%d min", ip, minutesWindow)
	var count int64
	timeWindow := time.Now().Add(-time.Duration(minutesWindow) * time.Minute)
	if err := r.db.Model(&models.GuestSession{}).
		Where("session_ip_address = ? AND last_attempt_at > ?", ip, timeWindow).
		Count(&count).Error; err != nil {
		log.Printf("[guestRepository.CountVerificationAttemptsByIP] ERROR: Failed to count attempts - ip=%s: %v", ip, err)
		return 0, err
	}
	log.Printf("[guestRepository.CountVerificationAttemptsByIP] INFO: Found %d attempts from IP %s", count, ip)
	return int(count), nil
}

func (r *guestRepository) DeleteExpiredSessions() (int64, error) {
	log.Printf("[guestRepository.DeleteExpiredSessions] INFO: Deleting expired sessions")
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&models.GuestSession{})
	if result.Error != nil {
		log.Printf("[guestRepository.DeleteExpiredSessions] ERROR: Failed to delete expired sessions: %v", result.Error)
		return 0, result.Error
	}
	log.Printf("[guestRepository.DeleteExpiredSessions] INFO: Deleted %d expired sessions", result.RowsAffected)
	return result.RowsAffected, nil
}
