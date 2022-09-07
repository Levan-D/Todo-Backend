package system

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

	system := r.Group("/system", auth.Authorization)
	{
		system.Get("/", h.getSystem)
	}
}

type systemResponse struct {
	IsVerified *bool `json:"is_verified"`
}

// @Tags System
// @Summary Retrieve a system
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} systemResponse
// @Failure 400 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /system [get]
func (h *handler) getSystem(c *fiber.Ctx) error {
	user := c.Locals(auth.LocalUser).(domain.User)

	return c.Status(http.StatusOK).JSON(systemResponse{
		IsVerified: user.IsVerified,
	})
}
