package utils

import "github.com/gofiber/fiber/v2"

func GetRealIP(c *fiber.Ctx) string {
	httpIP := c.IP()
	headerIP := string(c.Request().Header.Peek("X-Remote-IP"))

	if headerIP != "" {
		return headerIP
	}
	return httpIP
}
