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
	UpdatePositionByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID, input UpdateTaskPositionInput) error
	DeleteByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID) error
}

type CreateTaskInput struct {
	Description string
}

type UpdateTaskInput struct {
	Description *string
	IsCompleted *bool
}

type UpdateTaskPositionInput struct {
	EndpointID uuid.UUID
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

	inputData := make(map[string]interface{})

	if input.Description != nil && *input.Description != "" {
		inputData["description"] = input.Description
	} else if input.Description != nil && *input.Description == "" {
		inputData["description"] = ""
	}

	if input.IsCompleted != nil {
		inputData["is_completed"] = input.IsCompleted

		if *input.IsCompleted == false {
			inputData["completed_at"] = nil
		} else {
			inputData["completed_at"] = utils.NewTimeNow()
		}
	}

	err := s.repository.UpdateByID(userId, listId, id, inputData)
	if err != nil {
		return domain.Task{}, err
	}

	return s.repository.FindByID(userId, listId, id)
}

func (s *service) UpdatePositionByID(userId uuid.UUID, listId uuid.UUID, id uuid.UUID, input UpdateTaskPositionInput) error {
	current, err := s.repository.FindByID(userId, listId, id)
	if err != nil {
		return err
	}

	endpoint, err := s.repository.FindByID(userId, listId, input.EndpointID)
	if err != nil {
		return err
	}

	if current.ID == endpoint.ID {
		return nil
	}

	lists, err := s.repository.FindAll(userId, listId)
	if err != nil {
		return err
	}

	isNextPosition := true
	if current.Position < endpoint.Position {
		isNextPosition = true
	} else if current.Position > endpoint.Position {
		isNextPosition = false
	} else {
		return errors.New("list position has equals")
	}

	for index, item := range lists {
		if item.ID == current.ID {
			lists, err = ArrayDelete(lists, index)
			if err != nil {
				return err
			}
			break
		}
	}

	for index, item := range lists {
		if item.ID == endpoint.ID {
			indexInc := index
			if isNextPosition {
				indexInc++
			}

			lists, err = ArrayInsert(lists, indexInc, current)
			if err != nil {
				return err
			}
			break
		}
	}

	for index, item := range lists {
		err = s.repository.UpdateByID(userId, listId, item.ID, map[string]interface{}{"position": int32(index + 1)})
		if err != nil {
			return err
		}
	}

	return nil
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

func ArrayInsert(origin []domain.Task, index int, value domain.Task) ([]domain.Task, error) {
	if index < 0 {
		return nil, errors.New("Index cannot be less than 0")
	}

	if index >= len(origin) {
		return append(origin, value), nil
	}

	origin = append(origin[:index+1], origin[index:]...)
	origin[index] = value

	return origin, nil
}

func ArrayDelete(origin []domain.Task, index int) ([]domain.Task, error) {
	if index < 0 || index >= len(origin) {
		return nil, errors.New("Index cannot be less than 0")
	}

	origin = append(origin[:index], origin[index+1:]...)

	return origin, nil
}
