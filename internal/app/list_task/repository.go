package list_task

import (
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

type Repository interface {
	FindAll(userId uuid.UUID, listId uuid.UUID) (lists []domain.Task, err error)
	FindByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID) (item domain.Task, err error)
	Create(data domain.Task) (uuid.UUID, error)
	UpdateByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID, data domain.Task) (err error)
	DeleteByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID) error
	VerifyUserListByID(userId uuid.UUID, listId uuid.UUID) bool
	GetLastElementPosition(listId uuid.UUID) int32
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r repository) FindAll(userId uuid.UUID, listId uuid.UUID) (lists []domain.Task, err error) {
	err = r.db.Where("list_id = ?", listId).Order("position asc").Find(&lists).Error
	if err != nil {
		return []domain.Task{}, err
	}

	return lists, nil
}

func (r *repository) FindByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID) (item domain.Task, err error) {
	err = r.db.Where("list_id = ?", listId).Where("id = ?", id).First(&item).Error
	return item, err
}

func (r repository) Create(data domain.Task) (uuid.UUID, error) {
	err := r.db.Create(&data).Error
	if err != nil {
		return uuid.UUID{}, err
	}

	return data.ID, nil
}

func (r repository) UpdateByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID, data domain.Task) (err error) {
	err = r.db.Where("list_id = ?", listId).Where("id = ?", id).Updates(&data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) DeleteByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID) error {
	err := r.db.Where("list_id = ?", listId).Where("id = ?", id).Delete(&domain.Task{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) VerifyUserListByID(userId uuid.UUID, listId uuid.UUID) bool {
	var item domain.List
	err := r.db.Where("user_id = ?", userId).Where("id = ?", listId).First(&item).Error
	if err != nil {
		return false
	}
	return true
}

func (r *repository) GetLastElementPosition(listId uuid.UUID) int32 {
	var lastElement domain.Task
	if err := r.db.Where("list_id = ?", listId).Order("position desc").First(&lastElement).Error; err != nil {
		return 0
	}
	return lastElement.Position
}
