package response

import (
	"github.com/Levan-D/Todo-Backend/internal/app/errors"
	"github.com/Levan-D/Todo-Backend/pkg/logger"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	"github.com/Levan-D/Todo-Backend/pkg/validator"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	ErrInvalidBody       = "Invalid body provided"
	ErrInvalidQuery      = "Invalid query provided"
	ErrForbidden         = "Invalid permissions"
	ErrInvalidValidation = "Cannot be validated input"
	ErrUnauthorized      = "Unauthorized"
	ErrUndetected        = "Undetected"
)

type Message struct {
	Message string `json:"message"`
}

type Error struct {
	Locale string `json:"locale" enums:"INVALID_BODY,INVALID_VALIDATION,INVALID_CREDENTIALS,INVALID_ID,INVALID_FILE,INVALID_HASH,INVALID_BIRTHDAY_DATE,NOT_FOUND,UNDEFINED,AUTH_EMAIL_NOT_FOUND,AUTH_FACEBOOK_FAILED,AUTH_GOOGLE_FAILED,AUTH_APPLE_FAILED,AUTH_RESET_TOKEN_NOT_FOUND,AUTH_RESET_TOKEN_EXPIRED,AUTH_PASSWORD_NOT_MATCH,AUTH_CONFIRMATION_CODE_NOT_FOUND,AUTH_CONFIRMATION_CODE_EXPIRED,PAYMENT_NOT_FOUND,PAYMENT_ALREADY_PAID,PAYMENT_ALREADY_CANCELLED,PAYMENT_RETURN_NOT_FOUND,PAYMENT_HISTORY_NOT_FOUND,PAYMENT_BANK_NOT_CREATED,PAYMENT_WITH_BANK_CARD_PROBLEM,PAYMENT_PROBLEM_BANK_RETURN,EVENT_VOUCHER_NOT_FOUND,EVENT_VOUCHER_HAS_CLAIM_LIMIT,QUESTIONNAIRE_ALREADY_ANSWERED,QUESTIONNAIRE_NOT_FOUND,QUESTIONNAIRE_QUESTION_NOT_FOUND,QUESTIONNAIRE_QUESTION_OPTION_NOT_FOUND,DOCUMENT_NOT_FOUND,PARKING_INVALID_CODE,PARKING_TICKET_CANT_EVALUATION,PARKING_SPOT_ALREADY_LINKED,PARKING_CANNOT_PAID,PARKING_SPOT_NOT_FOUND,PARKING_PAYMENT_NOT_POSSIBLE,PARKING_PAYMENT_ZERO_AMOUNT,PARKING_PAYMENT_ALREADY_PAID,PARKING_GATEWAY_PROBLEM,SCAN_NOT_FOUND,LEVEL_NOT_FOUND,TENANT_NOT_FOUND,OFFER_NOT_FOUND,EVENT_NOT_FOUND,EVENT_REGISTRATION_ENDED,EVENT_TICKET_NOT_FOUND,EVENT_TICKETS_NOT_FOUND,EVENT_TICKETS_SHARE_INVALID_COUNT,EVENT_TICKETS_SHARE_NOT_FOUND,USER_NOT_FOUND,USER_ALREADY_VERIFIED,USER_NOT_VERIFIED,USER_DEACTIVATED,USER_BLOCKED,USER_BANK_CARD_NOT_FOUND,USER_BANK_CARDS_NOT_FOUND"`
	Error  string `json:"error"`
}

func NewError(c *fiber.Ctx, err error) error {
	statusCode := errors.GetStatusCode(err)
	errorLocale := errors.GetLocale(err)

	logFields := log.Fields{
		"request_id":   string(c.Response().Header.Peek("X-Request-Id")),
		"status_code":  statusCode,
		"method":       string(c.Request().Header.Method()),
		"path":         string(c.Request().URI().Path()),
		"query":        string(c.Request().URI().QueryArgs().QueryString()),
		"body":         string(c.Request().Body()),
		"error":        err,
		"error-locale": errorLocale,
		"ip":           utils.GetRealIP(c),
	}

	if statusCode == http.StatusInternalServerError {
		logger.Error(logFields, err.Error())
	} else {
		logger.Warn(logFields, err.Error())
	}

	return c.Status(statusCode).JSON(Error{
		Locale: string(errorLocale),
		Error:  err.Error(),
	})
}

type ErrorValidator struct {
	Locale string                    `json:"locale"`
	Error  string                    `json:"error"`
	Fields []*validator.ValidateItem `json:"fields"`
}

func NewErrorValidator(c *fiber.Ctx, err error, validateErrors []*validator.ValidateItem) error {
	statusCode := errors.GetStatusCode(err)
	errorLocale := errors.GetLocale(err)

	logFields := log.Fields{
		"request_id":      string(c.Response().Header.Peek("X-Request-Id")),
		"status_code":     statusCode,
		"method":          string(c.Request().Header.Method()),
		"path":            string(c.Request().URI().Path()),
		"query":           string(c.Request().URI().QueryArgs().QueryString()),
		"body":            string(c.Request().Body()),
		"ip":              utils.GetRealIP(c),
		"error":           err,
		"validate-errors": validateErrors,
	}

	if statusCode == http.StatusInternalServerError {
		logger.Error(logFields, err.Error())
	} else {
		logger.Warn(logFields, err.Error())
	}

	return c.Status(statusCode).JSON(ErrorValidator{
		Locale: string(errorLocale),
		Error:  err.Error(),
		Fields: validateErrors,
	})
}
