package repositories

import (
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	FindAll(limit, offset int) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	log.Printf("[userRepository.Create] INFO: Creating user - email=%s, role=%s", user.Email, user.Role)
	if err := r.db.Create(user).Error; err != nil {
		log.Printf("[userRepository.Create] ERROR: Failed to create user - email=%s: %v", user.Email, err)
		return err
	}
	log.Printf("[userRepository.Create] INFO: User created successfully - userID=%d, email=%s", user.ID, user.Email)
	return nil
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	log.Printf("[userRepository.FindByID] INFO: Retrieving user - userID=%d", id)
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		log.Printf("[userRepository.FindByID] ERROR: Failed to find user - userID=%d: %v", id, err)
		return nil, err
	}
	log.Printf("[userRepository.FindByID] INFO: User found - userID=%d, email=%s", id, user.Email)
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	log.Printf("[userRepository.FindByEmail] INFO: Retrieving user - email=%s", email)
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("[userRepository.FindByEmail] ERROR: Failed to find user - email=%s: %v", email, err)
		return nil, err
	}
	log.Printf("[userRepository.FindByEmail] INFO: User found - userID=%d, email=%s", user.ID, email)
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	log.Printf("[userRepository.Update] INFO: Updating user - userID=%d, email=%s", user.ID, user.Email)
	if err := r.db.Save(user).Error; err != nil {
		log.Printf("[userRepository.Update] ERROR: Failed to update user - userID=%d: %v", user.ID, err)
		return err
	}
	log.Printf("[userRepository.Update] INFO: User updated successfully - userID=%d", user.ID)
	return nil
}

func (r *userRepository) Delete(id uint) error {
	log.Printf("[userRepository.Delete] INFO: Deleting user - userID=%d", id)
	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		log.Printf("[userRepository.Delete] ERROR: Failed to delete user - userID=%d: %v", id, err)
		return err
	}
	log.Printf("[userRepository.Delete] INFO: User deleted successfully - userID=%d", id)
	return nil
}

func (r *userRepository) FindAll(limit, offset int) ([]models.User, error) {
	log.Printf("[userRepository.FindAll] INFO: Listing users - limit=%d, offset=%d", limit, offset)
	var users []models.User
	if err := r.db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		log.Printf("[userRepository.FindAll] ERROR: Failed to list users: %v", err)
		return nil, err
	}
	log.Printf("[userRepository.FindAll] INFO: Users listed - count=%d", len(users))
	return users, nil
}
