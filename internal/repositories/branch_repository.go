package repositories

import (
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type BranchRepository interface {
	Create(branch *models.Branch) error
	FindByID(id uint) (*models.Branch, error)
	FindByCode(code string) (*models.Branch, error)
	Update(branch *models.Branch) error
	Delete(id uint) error
	FindAll(onlyActive, onlyPickup bool) ([]models.Branch, error)
}

type branchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(branch *models.Branch) error {
	log.Printf("[branchRepository.Create] INFO: Creating branch - code=%s, name=%s", branch.Code, branch.Name)
	if err := r.db.Create(branch).Error; err != nil {
		log.Printf("[branchRepository.Create] ERROR: code=%s: %v", branch.Code, err)
		return err
	}
	return nil
}

func (r *branchRepository) FindByID(id uint) (*models.Branch, error) {
	var b models.Branch
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *branchRepository) FindByCode(code string) (*models.Branch, error) {
	var b models.Branch
	if err := r.db.Where("code = ?", code).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *branchRepository) Update(branch *models.Branch) error {
	log.Printf("[branchRepository.Update] INFO: Updating branch - branchID=%d", branch.ID)
	return r.db.Save(branch).Error
}

func (r *branchRepository) Delete(id uint) error {
	log.Printf("[branchRepository.Delete] INFO: Deleting branch - branchID=%d", id)
	return r.db.Delete(&models.Branch{}, id).Error
}

func (r *branchRepository) FindAll(onlyActive, onlyPickup bool) ([]models.Branch, error) {
	var branches []models.Branch
	q := r.db.Model(&models.Branch{})
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	if onlyPickup {
		q = q.Where("is_pickup_point = ?", true)
	}
	if err := q.Order("id ASC").Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}
