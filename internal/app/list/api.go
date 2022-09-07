package list

import (
	"github.com/Levan-D/Todo-Backend/internal/app/auth"
	"github.com/Levan-D/Todo-Backend/internal/app/errors"
	"github.com/Levan-D/Todo-Backend/internal/app/response"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/Levan-D/Todo-Backend/pkg/validator"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
)

type handler struct {
	service Service
}

func RegisterHandlers(r fiber.Router, service Service) {
	h := handler{service}

	route := r.Group("/lists", auth.Authorization)
	{
		route.Get("/", h.getListAll)
		route.Post("/", h.createList)
		route.Patch("/:id", h.updateListById)
		route.Put("/:id/position", h.updateListPositionById)
		route.Delete("/:id", h.deleteListById)
	}
}

// @Tags Lists
// @Summary List all lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []domain.List
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists [get]
func (h *handler) getListAll(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	result, err := h.service.GetAll(user.ID)
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusOK).JSON(result)
}

type createListInput struct {
	Title string `json:"title" validate:"required"`
}

// @Tags Lists
// @Summary Create a list
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body createListInput true "Input data"
// @Success 200 {object} domain.List
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists [post]
func (h *handler) createList(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	var input createListInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, err)
	}

	errValidation := validator.Validate(&input)
	if errValidation != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidation)
	}

	res, err := h.service.Create(user.ID, CreateListInput{
		Title: input.Title,
	})
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

type updateListInput struct {
	Title      *string    `json:"title"`
	Color      *string    `json:"color"`
	ReminderAt *time.Time `json:"reminder_at"`
}

// @Tags Lists
// @Summary Update a list
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID"
// @Param input body updateListInput true "Input data"
// @Success 200 {object} domain.List
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{id} [patch]
func (h *handler) updateListById(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	id, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	var input updateListInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, err)
	}

	res, err := h.service.UpdateByID(user.ID, id, UpdateListInput{
		Title:      input.Title,
		Color:      input.Color,
		ReminderAt: input.ReminderAt,
	})
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

type updateListPositionInput struct {
	NextPosition int32 `json:"next_position"`
}

// @Tags Lists
// @Summary Update a list position
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID"
// @Param input body updateListPositionInput true "Input data"
// @Success 200 {object} response.Message
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{id}/position [put]
func (h *handler) updateListPositionById(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	id, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	var input updateListPositionInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, err)
	}

	err = h.service.UpdatePositionByID(user.ID, id, UpdateListPositionInput{
		NextPosition: input.NextPosition,
	})
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusOK).JSON(response.Message{Message: "position successfully changed"})
}

// @Tags Lists
// @Summary Delete a list
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID"
// @Success 204 ""
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /lists/{id} [delete]
func (h *handler) deleteListById(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	id, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrInvalidID, errors.LocaleInvalidID))
	}

	err = h.service.DeleteByID(user.ID, id)
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusNoContent).JSON(nil)
}
