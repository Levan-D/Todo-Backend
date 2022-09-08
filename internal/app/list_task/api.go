package list_task

import (
	"github.com/Levan-D/Todo-Backend/internal/app/auth"
	"github.com/Levan-D/Todo-Backend/internal/app/errors"
	"github.com/Levan-D/Todo-Backend/internal/app/response"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/Levan-D/Todo-Backend/pkg/validator"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type handler struct {
	service Service
}

func RegisterHandlers(r fiber.Router, service Service) {
	h := handler{service}

	route := r.Group("/lists/:listId/tasks", auth.Authorization)
	{
		route.Get("/", h.getTaskAll)
		route.Post("/", h.createTask)
		route.Patch("/:id", h.updateTaskById)
		route.Put("/:id/position", h.updateTaskPositionById)
		route.Delete("/:id", h.deleteTaskById)
	}
}

// @Tags Tasks
// @Summary List all tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []domain.Task
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{listId}/tasks [get]
func (h *handler) getTaskAll(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	listId, err := uuid.FromString(c.Params("listId"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	result, err := h.service.GetAll(user.ID, listId)
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusOK).JSON(result)
}

type createListInput struct {
	Description string `json:"description" validate:"required"`
}

// @Tags Tasks
// @Summary Create a task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body createListInput true "Input data"
// @Success 200 {object} domain.Task
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{listId}/tasks [post]
func (h *handler) createTask(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	listId, err := uuid.FromString(c.Params("listId"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	var input createListInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, err)
	}

	errValidation := validator.Validate(&input)
	if errValidation != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidation)
	}

	res, err := h.service.Create(user.ID, listId, CreateTaskInput{
		Description: input.Description,
	})
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

type updateListInput struct {
	Description *string `json:"description"`
	IsCompleted *bool   `json:"is_completed"`
}

// @Tags Tasks
// @Summary Update a task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID"
// @Param input body updateListInput true "Input data"
// @Success 200 {object} domain.Task
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{listId}/tasks/{id} [patch]
func (h *handler) updateTaskById(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	listId, err := uuid.FromString(c.Params("listId"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	id, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	var input updateListInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, err)
	}

	res, err := h.service.UpdateByID(user.ID, listId, id, UpdateTaskInput{
		Description: input.Description,
		IsCompleted: input.IsCompleted,
	})
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

type updateTaskPositionInput struct {
	EndpointID uuid.UUID `json:"endpoint_id" validate:"required"`
}

// @Tags Tasks
// @Summary Update a task position
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID"
// @Param input body updateTaskPositionInput true "Input data"
// @Success 200 {object} response.Message
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{listId}/tasks/{id}/position [put]
func (h *handler) updateTaskPositionById(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	listId, err := uuid.FromString(c.Params("listId"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	id, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	var input updateTaskPositionInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, err)
	}

	err = h.service.UpdatePositionByID(user.ID, listId, id, UpdateTaskPositionInput{
		EndpointID: input.EndpointID,
	})
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusOK).JSON(response.Message{Message: "position successfully changed"})
}

// @Tags Tasks
// @Summary Delete a task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID"
// @Success 204 ""
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{listId}/tasks/{id} [delete]
func (h *handler) deleteTaskById(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	listId, err := uuid.FromString(c.Params("listId"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	id, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	err = h.service.DeleteByID(user.ID, listId, id)
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusNoContent).JSON(nil)
}
