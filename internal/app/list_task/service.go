package list_task

import (
	"github.com/Levan-D/Todo-Backend/internal/app/errors"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	uuid "github.com/satori/go.uuid"
)

type service struct {
	repository Repository
}

type Service interface {
	GetAll(userId uuid.UUID, listId uuid.UUID) ([]domain.Task, error)
	Create(userId uuid.UUID, listId uuid.UUID, input CreateTaskInput) (domain.Task, error)
	UpdateByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID, input UpdateTaskInput) (domain.Task, error)
	DeleteByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID) error
}

type CreateTaskInput struct {
	Description string
}

type UpdateTaskInput struct {
	Description *string
	IsCompleted *bool
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) GetAll(userId uuid.UUID, listId uuid.UUID) ([]domain.Task, error) {
	if ok := s.repository.VerifyUserListByID(userId, listId); !ok {
		return []domain.Task{}, errors.New("user have not permission on this list")
	}

	return s.repository.FindAll(userId, listId)
}

func (s *service) Create(userId uuid.UUID, listId uuid.UUID, input CreateTaskInput) (domain.Task, error) {
	if ok := s.repository.VerifyUserListByID(userId, listId); !ok {
		return domain.Task{}, errors.New("user have not permission on this list")
	}

	id, err := s.repository.Create(domain.Task{
		ListID:      &listId,
		Description: input.Description,
		Position:    s.repository.GetLastElementPosition(listId) + 1,
		IsCompleted: utils.NewFalse(),
	})
	if err != nil {
		return domain.Task{}, err
	}

	return s.repository.FindByID(userId, listId, id)
}

func (s *service) UpdateByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID, input UpdateTaskInput) (domain.Task, error) {
	if ok := s.repository.VerifyUserListByID(userId, listId); !ok {
		return domain.Task{}, errors.New("user have not permission on this list")
	}

	inputData := domain.Task{}

	if input.Description != nil && *input.Description != "" {
		inputData.Description = *input.Description
	}

	if input.IsCompleted != nil {
		inputData.IsCompleted = input.IsCompleted

		if *input.IsCompleted == false {
			inputData.CompletedAt = nil
		} else {
			inputData.CompletedAt = utils.NewTimeNow()
		}
	}

	err := s.repository.UpdateByID(userId, listId, id, inputData)
	if err != nil {
		return domain.Task{}, err
	}

	return s.repository.FindByID(userId, listId, id)
}

func (s *service) DeleteByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID) error {
	if ok := s.repository.VerifyUserListByID(userId, listId); !ok {
		return errors.New("user have not permission on this list")
	}

	err := s.repository.DeleteByID(userId, listId, id)
	if err != nil {
		return err
	}
	return nil
}
