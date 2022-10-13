package list

import (
	"errors"
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
	EndpointID uuid.UUID
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
	inputData := make(map[string]interface{})

	if input.Color != nil && *input.Color != "" {
		inputData["color"] = input.Color
	} else if input.Color != nil && *input.Color == "" {
		inputData["color"] = nil
	}

	if input.Title != nil && *input.Title != "" {
		inputData["title"] = *input.Title
	} else if input.Title != nil && *input.Title == "" {
		inputData["title"] = nil
	}

	zeroTime := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	if input.ReminderAt != nil && !zeroTime.Equal(*input.ReminderAt) {
		inputData["reminder_at"] = input.ReminderAt
		inputData["is_reminded"] = utils.NewFalse()
	} else if input.ReminderAt != nil && zeroTime.Equal(*input.ReminderAt) {
		inputData["reminder_at"] = nil
		inputData["is_reminded"] = utils.NewFalse()
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

	endpoint, err := s.repository.FindByID(userId, input.EndpointID)
	if err != nil {
		return err
	}

	if current.ID == endpoint.ID {
		return nil
	}

	lists, err := s.repository.FindAll(userId)
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
		err = s.repository.UpdateByID(userId, item.ID, map[string]interface{}{"position": int32(index + 1)})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) DeleteByID(userId uuid.UUID, id uuid.UUID) error {
	err := s.repository.DeleteByID(userId, id)
	if err != nil {
		return err
	}
	return nil
}

func ArrayInsert(origin []domain.List, index int, value domain.List) ([]domain.List, error) {
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

func ArrayDelete(origin []domain.List, index int) ([]domain.List, error) {
	if index < 0 || index >= len(origin) {
		return nil, errors.New("Index cannot be less than 0")
	}

	origin = append(origin[:index], origin[index+1:]...)

	return origin, nil
}
