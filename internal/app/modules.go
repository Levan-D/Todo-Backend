package app

import (
	"github.com/Levan-D/Todo-Backend/internal/app/auth"
	"github.com/Levan-D/Todo-Backend/internal/app/health"
	"github.com/Levan-D/Todo-Backend/internal/app/list"
	"github.com/Levan-D/Todo-Backend/internal/app/list_task"
	"github.com/Levan-D/Todo-Backend/internal/app/me"
	"github.com/Levan-D/Todo-Backend/internal/app/profile"
	"github.com/Levan-D/Todo-Backend/internal/app/system"
	"github.com/Levan-D/Todo-Backend/pkg/argon2id"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Initialize(app *fiber.App, db *gorm.DB) {
	argon2 := argon2id.NewArgon2ID()

	v1 := app.Group("/api/v1", func(c *fiber.Ctx) error {
		c.Set("Version", "v1")
		return c.Next()
	})
	{
		auth.RegisterHandlers(v1, auth.NewService(auth.NewRepository(db), argon2))
		list.RegisterHandlers(v1, list.NewService(list.NewRepository(db)))
		list_task.RegisterHandlers(v1, list_task.NewService(list_task.NewRepository(db)))
		profile.RegisterHandlers(v1, profile.NewService(profile.NewRepository(db)))
		me.RegisterHandlers(v1, me.NewService(me.NewRepository(db)))
		system.RegisterHandlers(v1, system.NewService(system.NewRepository(db)))
		health.RegisterHandlers(v1)
	}
}
