package auth

import (
	"github.com/Levan-D/Todo-Backend/internal/app/errors"
	"github.com/Levan-D/Todo-Backend/internal/app/response"
	"github.com/Levan-D/Todo-Backend/internal/common"
	"github.com/Levan-D/Todo-Backend/pkg/auth"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/gofiber/fiber/v2"
)

const (
	LocalUser = "USER"
)

// Examples:

// metadata, err := auth.ExtractTokenMetadata(c.Request())
// if err != nil {
//	 return response.NewError(c, http.StatusUnauthorized, response.ErrUnauthorized, err)
// }

// user := c.Locals(LocalUser).(domain.User)
// fmt.Println("User: ", user)

// settings := c.Locals(LocalSetting).(service.SettingResponse)
// fmt.Println("Settings: ", settings.GeneralLimitOfferInLevel)

func Authorization(c *fiber.Ctx) error {
	err := auth.TokenValid(c.Request())
	if err != nil {
		return response.NewError(c, errors.StatusUnauthorized.New(errors.ErrUnauthorized))
	}

	meta, err := auth.ExtractTokenMetadata(c.Request())
	if err != nil {
		return response.NewError(c, errors.StatusUnauthorized.New(errors.ErrUnauthorized))
	}

	user, err := common.FindUserByID(meta.UserID)
	if err != nil {
		return response.NewError(c, errors.StatusUnauthorized.New(errors.ErrUnauthorized))
	}

	c.Locals(LocalUser, user)

	return c.Next()
}

func Verification(c *fiber.Ctx) error {
	user := c.Locals(LocalUser).(domain.User)
	if *user.IsVerified != true {
		return response.NewError(c, errors.StatusNotAcceptable.LocaleNew("user has not verified", errors.LocaleUserNotVerified))
	}

	return c.Next()
}
