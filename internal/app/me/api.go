package me

import (
	"github.com/Levan-D/Todo-Backend/internal/app/auth"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type handler struct {
	service Service
}

func RegisterHandlers(r fiber.Router, service Service) {
	h := handler{service}

	profile := r.Group("/me", auth.Authorization)
	{
		profile.Get("/", h.getMe)
	}
}

type meResponse struct {
	Avatar    *string `json:"avatar"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
}

// @Tags Me
// @Summary Retrieve a me
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} meResponse
// @Failure 400 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /me [get]
func (h *handler) getMe(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	return c.Status(http.StatusOK).JSON(meResponse{
		Avatar:    user.Avatar,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
}
