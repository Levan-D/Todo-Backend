package list

import (
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	uuid "github.com/satori/go.uuid"
	"time"
)

type service struct {
	repository Repository
}

type Service interface {
	GetAll(userId uuid.UUID) ([]domain.List, error)
	Create(userId uuid.UUID, input CreateListInput) (domain.List, error)
	UpdateByID(userId uuid.UUID, id uuid.UUID, input UpdateListInput) (domain.List, error)
	UpdatePositionByID(userId uuid.UUID, id uuid.UUID, input UpdateListPositionInput) error
	DeleteByID(userId uuid.UUID, id uuid.UUID) error
}

type CreateListInput struct {
	Title string
}

type UpdateListInput struct {
	Title      *string
	Color      *string
	ReminderAt *time.Time
}
type UpdateListPositionInput struct {
	NextPosition int32
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) GetAll(userId uuid.UUID) ([]domain.List, error) {
	return s.repository.FindAll(userId)
}

func (s *service) Create(userId uuid.UUID, input CreateListInput) (domain.List, error) {
	id, err := s.repository.Create(domain.List{
		UserID:   &userId,
		Title:    input.Title,
		Position: s.repository.GetLastElementPosition(userId) + 1,
	})
	if err != nil {
		return domain.List{}, err
	}

	return s.repository.FindByID(userId, id)
}

func (s *service) UpdateByID(userId uuid.UUID, id uuid.UUID, input UpdateListInput) (domain.List, error) {
	inputData := domain.List{}

	if input.Color != nil && *input.Color != "" {
		inputData.Color = input.Color
	}

	if input.Title != nil && *input.Title != "" {
		inputData.Title = *input.Title
	}

	if input.ReminderAt != nil {
		inputData.ReminderAt = input.ReminderAt
		inputData.IsReminded = utils.NewFalse()
	}

	err := s.repository.UpdateByID(userId, id, inputData)
	if err != nil {
		return domain.List{}, err
	}

	return s.repository.FindByID(userId, id)
}

func (s *service) UpdatePositionByID(userId uuid.UUID, id uuid.UUID, input UpdateListPositionInput) error {
	current, err := s.repository.FindByID(userId, id)
	if err != nil {
		return err
	}

	next, err := s.repository.FindByPosition(userId, input.NextPosition)
	if err != nil {
		return err
	}

	if current.ID == next.ID {
		return nil
	}

	// TODO: finish pos

	return nil
}

func (s *service) DeleteByID(userId uuid.UUID, id uuid.UUID) error {
	err := s.repository.DeleteByID(userId, id)
	if err != nil {
		return err
	}
	return nil
}
