package health

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type handler struct {
}

func RegisterHandlers(r fiber.Router) {
	h := handler{}

	route := r.Group("/health")
	{
		route.Get("/", h.getHealth)
	}
}

// @Tags Scan
// @Summary Retrieve a url details
// @Accept json
// @Produce json
// @Success 200 ""
// @Router /health [get]
func (h *handler) getHealth(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusOK)
}
