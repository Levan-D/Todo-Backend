package me

import (
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

type Repository interface {
	FindByID(id uuid.UUID) (domain.User, error)
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByID(id uuid.UUID) (domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return user, err
}
