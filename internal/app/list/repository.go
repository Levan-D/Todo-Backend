package list

import (
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

type Repository interface {
	FindAll(userId uuid.UUID) (lists []domain.List, err error)
	FindByID(userId uuid.UUID, id uuid.UUID) (item domain.List, err error)
	FindByPosition(userId uuid.UUID, position int32) (item domain.List, err error)
	Create(data domain.List) (uuid.UUID, error)
	UpdateByID(userId uuid.UUID, id uuid.UUID, data map[string]interface{}) (err error)
	DeleteByID(userId uuid.UUID, id uuid.UUID) error
	GetLastElementPosition(userId uuid.UUID) int32
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r repository) FindAll(userId uuid.UUID) (lists []domain.List, err error) {
	err = r.db.Where("user_id = ?", userId).Order("position asc").Find(&lists).Error
	if err != nil {
		return []domain.List{}, err
	}

	return lists, nil
}

func (r *repository) FindByID(userId uuid.UUID, id uuid.UUID) (item domain.List, err error) {
	err = r.db.Where("user_id = ?", userId).Where("id = ?", id).First(&item).Error
	return item, err
}

func (r *repository) FindByPosition(userId uuid.UUID, position int32) (item domain.List, err error) {
	err = r.db.Where("user_id = ?", userId).Where("position = ?", position).First(&item).Error
	return item, err
}

func (r repository) Create(data domain.List) (uuid.UUID, error) {
	err := r.db.Create(&data).Error
	if err != nil {
		return uuid.UUID{}, err
	}

	return data.ID, nil
}

func (r repository) UpdateByID(userId uuid.UUID, id uuid.UUID, data map[string]interface{}) (err error) {
	err = r.db.Model(&domain.List{}).Where("user_id = ?", userId).Where("id = ?", id).Updates(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) DeleteByID(userId uuid.UUID, id uuid.UUID) error {
	err := r.db.Where("user_id = ?", userId).Where("id = ?", id).Delete(&domain.List{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetLastElementPosition(userId uuid.UUID) int32 {
	var lastElement domain.List
	if err := r.db.Where("user_id = ?", userId).Order("position desc").First(&lastElement).Error; err != nil {
		return 0
	}
	return lastElement.Position
}
