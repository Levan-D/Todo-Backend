package auth

import (
	"fmt"
	"github.com/Levan-D/Todo-Backend/internal/app/errors"
	"github.com/Levan-D/Todo-Backend/internal/app/response"
	"github.com/Levan-D/Todo-Backend/pkg/auth"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	"github.com/Levan-D/Todo-Backend/pkg/validator"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type handler struct {
	service Service
}

func RegisterHandlers(r fiber.Router, service Service) {
	h := handler{service}

	route := r.Group("/auth")
	{
		route.Post("/login", h.login)
		route.Post("/sign-up", h.signUp)
		route.Post("/logout", h.logout)

		route.Post("/forgot", h.forgotPassword)
		route.Post("/forgot/confirm", h.forgotConfirm)
		route.Post("/forgot/reset", h.forgotResetPassword)

		route.Post("/refresh", h.refresh)
	}
}

type loginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// @Tags Auth
// @Summary Login
// @Accept json
// @Produce json
// @Param input body loginInput true "login input"
// @Success 200 {object} tokenResponse
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/login [post]
func (h *handler) login(c *fiber.Ctx) error {
	var input loginInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrParseBody, errors.LocaleInvalidBody))
	}

	errValidation := validator.Validate(&input)
	if errValidation != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidation)
	}

	user, err := h.service.Login(input.Email, input.Password, utils.GetRealIP(c), string(c.Request().Header.Peek("User-Agent")))
	if err != nil {
		return response.NewError(c, err)
	}

	createToken, err := auth.CreateToken(user.ID)
	if err != nil {
		return response.NewError(c, errors.StatusInternalServer.LocaleWrapf(err, errors.ErrFailedCreateToken, errors.LocaleUndefined))
	}

	saveErr := auth.CreateAuth(user.ID, createToken)
	if saveErr != nil {
		return response.NewError(c, errors.StatusInternalServer.LocaleWrapf(saveErr, errors.ErrFailedCreateToken, errors.LocaleUndefined))
	}

	return c.Status(http.StatusOK).JSON(tokenResponse{
		AccessToken:  createToken.AccessToken,
		RefreshToken: createToken.RefreshToken,
	})
}

type logoutInput struct {
	PushToken string `json:"push_token"`
}

// @Tags Auth
// @Summary Logout
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body logoutInput false "logout input"
// @Success 200 {object} response.Message
// @Failure 401 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/logout [post]
func (h *handler) logout(c *fiber.Ctx) error {
	metadata, err := auth.ExtractTokenMetadata(c.Request())
	if err != nil {
		return response.NewError(c, errors.StatusUnauthorized.LocaleWrapf(err, errors.ErrUnauthorized, errors.LocaleUndefined))
	}

	deleteError := auth.DeleteTokens(metadata)
	if deleteError != nil {
		//return response.NewError(c, http.StatusInternalServerError, "User can not be logout", deleteError)
	}

	var input logoutInput
	if err := c.BodyParser(&input); err == nil {

	}

	return c.Status(http.StatusOK).JSON(response.Message{Message: "Successfully logged out"})
}

type signUpInput struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

// @Tags Auth
// @Summary Sign Up
// @Accept json
// @Produce json
// @Param input body signUpInput true "sign up input"
// @Success 200 {object} tokenResponse
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/sign-up [post]
func (h *handler) signUp(c *fiber.Ctx) error {
	var input signUpInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrParseBody, errors.LocaleInvalidBody))
	}

	errValidate := validator.Validate(&input)
	if errValidate != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidate)
	}

	userId, err := h.service.SingUp(SignUpInput{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  input.Password,
	})
	if err != nil {
		return response.NewError(c, err)
	}

	createToken, err := auth.CreateToken(userId)
	if err != nil {
		return response.NewError(c, errors.StatusInternalServer.LocaleWrapf(err, errors.ErrFailedCreateToken, errors.LocaleUndefined))
	}

	saveErr := auth.CreateAuth(userId, createToken)
	if saveErr != nil {
		return response.NewError(c, errors.StatusInternalServer.LocaleWrapf(err, errors.ErrFailedCreateToken, errors.LocaleUndefined))
	}

	return c.Status(http.StatusOK).JSON(tokenResponse{
		AccessToken:  createToken.AccessToken,
		RefreshToken: createToken.RefreshToken,
	})
}

type forgotPasswordInput struct {
	Email string `json:"email" validate:"required,email"`
}

// @Tags Auth
// @Summary Forgot Password
// @Accept json
// @Produce json
// @Param input body forgotPasswordInput true "forgot password input"
// @Success 200 {object} response.Message
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/forgot [post]
func (h *handler) forgotPassword(c *fiber.Ctx) error {
	var input forgotPasswordInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrParseBody, errors.LocaleInvalidBody))
	}

	errValidate := validator.Validate(&input)
	if errValidate != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidate)
	}

	// generate password reset token 30 hours
	err := h.service.ForgotPassword(input.Email)
	if err != nil {
		return response.NewError(c, err)
	}

	return c.Status(http.StatusOK).JSON(response.Message{Message: "Recovery link successfully send to your email."})
}

type forgotConfirmInput struct {
	ConfirmationCode string `json:"confirmation_code" validate:"required"`
}

type forgotConfirmResponse struct {
	ConfirmationCode string `json:"confirmation_code"`
	Message          string `json:"message"`
}

// @Tags Auth
// @Summary Forgot Confirm by Code
// @Accept json
// @Produce json
// @Param input body forgotConfirmInput true "forgot confirm input"
// @Success 200 {object} forgotConfirmResponse
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/forgot/confirm [post]
func (h *handler) forgotConfirm(c *fiber.Ctx) error {
	var input forgotConfirmInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrParseBody, errors.LocaleInvalidBody))
	}

	errValidate := validator.Validate(&input)
	if errValidate != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidate)
	}

	//confirmationCode, err := h.service.CheckConfirmationCode(input.ConfirmationCode)
	//if err != nil {
	//	return response.NewError(c, err)
	//}

	return c.Status(http.StatusOK).JSON(forgotConfirmResponse{
		//ConfirmationCode: confirmationCode,
		Message: "Confirmation code is validated",
	})
}

type resetPasswordInput struct {
	ConfirmationCode string `json:"confirmation_code" validate:"required"`
	NewPassword      string `json:"new_password" validate:"required"`
	ConfirmPassword  string `json:"confirm_password" validate:"required"`
}

// @Tags Auth
// @Summary Reset Password
// @Accept json
// @Produce json
// @Param input body resetPasswordInput true "reset password input"
// @Success 200 {object} response.Message
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/forgot/reset [post]
func (h *handler) forgotResetPassword(c *fiber.Ctx) error {
	var input resetPasswordInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrParseBody, errors.LocaleInvalidBody))
	}

	errValidator := validator.Validate(&input)
	if errValidator != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidator)
	}

	//confirmationCode, err := h.service.CheckConfirmationCode(input.ConfirmationCode)
	//if err != nil {
	//	return response.NewError(c, err)
	//}

	//err = h.service.ResetPassword(confirmationCode, input.NewPassword, input.ConfirmPassword)
	//if err != nil {
	//	return response.NewError(c, err)
	//}

	return c.Status(http.StatusOK).JSON(response.Message{Message: "Password has been changed"})
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// @Tags Auth
// @Summary Refresh Token
// @Accept json
// @Produce json
// @Param input body refreshInput true "refresh token input"
// @Success 200 {object} tokenResponse
// @Failure 400 {object} response.Error
// @Failure 401 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/refresh [post]
func (h *handler) refresh(c *fiber.Ctx) error {
	var input refreshInput
	if err := c.BodyParser(&input); err != nil {
		return response.NewError(c, errors.StatusBadRequest.LocaleWrapf(err, errors.ErrParseBody, errors.LocaleInvalidBody))
	}

	errValidate := validator.Validate(&input)
	if errValidate != nil {
		return response.NewErrorValidator(c, errors.StatusBadRequest.LocaleNew(errors.ErrInvalidValidation, errors.LocaleInvalidValidation), errValidate)
	}

	token, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Get().JWT.RefreshSecret), nil
	})
	if err != nil {
		return response.NewError(c, errors.StatusUnauthorized.LocaleWrapf(err, "refresh token expired", errors.LocaleInvalidBody))
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return response.NewError(c, errors.StatusUnauthorized.LocaleNew("cannot be claim data", errors.LocaleInvalidBody))
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUuid := claims["refresh_uuid"].(string)

		deleted, delErr := auth.DeleteAuth(refreshUuid)
		if delErr != nil || deleted == 0 {
			return response.NewError(c, errors.StatusUnauthorized.LocaleWrapf(err, errors.ErrUnauthorized, errors.LocaleUndefined))
		}

		userId, _ := uuid.FromString(claims["user_id"].(string))

		createToken, createErr := auth.CreateToken(userId)
		if createErr != nil {
			return response.NewError(c, errors.StatusInternalServer.LocaleWrapf(createErr, errors.ErrFailedCreateToken, errors.LocaleUndefined))
		}

		saveErr := auth.CreateAuth(userId, createToken)
		if saveErr != nil {
			return response.NewError(c, errors.StatusInternalServer.LocaleWrapf(saveErr, errors.ErrFailedCreateToken, errors.LocaleUndefined))
		}

		return c.Status(http.StatusOK).JSON(tokenResponse{
			AccessToken:  createToken.AccessToken,
			RefreshToken: createToken.RefreshToken,
		})
	} else {
		return response.NewError(c, errors.StatusUnauthorized.LocaleWrapf(err, "refresh token expired", errors.LocaleUndefined))
	}
}
