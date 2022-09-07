package errors

import (
	"fmt"
	m_errors "github.com/pkg/errors"
	"net/http"
)

/*

return response.NewError(c, errors.StatusUnauthorized.New(errors.ErrUnauthorized))
errors.StatusNotFound.LocaleWrapf(err, "document not found", errors.LocaleNotFound)

errors.StatusInternalServer.LocaleWrapf(err, "cannot be parsed birthday", errors.LocaleUndefined)

if errors.Is(err, gorm.ErrRecordNotFound) {
	return EventGifts{}, errors.StatusNotFound.LocaleWrapf(err, "retrieve a event gifts", errors.LocaleNotFound)
}
return EventGifts{}, errors.StatusInternalServer.LocaleWrapf(err, "retrieve a event gifts", errors.LocaleUndefined)

*/

const (
	ErrUnauthorized       = "unauthorized"
	ErrParseBody          = "cannot be parse body"
	ErrInvalidID          = "cannot be parse id"
	ErrInvalidValidation  = "cannot be validated input"
	ErrInvalidFile        = "cannot be parse file"
	ErrFailedLoadSettings = "failed to load settings"
	ErrFailedCreateToken  = "tokens can not be created"
)

type ErrorLocale string

const (
	LocaleInvalidBody                  ErrorLocale = "INVALID_BODY"
	LocaleInvalidValidation            ErrorLocale = "INVALID_VALIDATION"
	LocaleInvalidCredentials           ErrorLocale = "INVALID_CREDENTIALS"
	LocaleInvalidID                    ErrorLocale = "INVALID_ID"
	LocaleInvalidFile                  ErrorLocale = "INVALID_FILE"
	LocaleInvalidHash                  ErrorLocale = "INVALID_HASH"
	LocaleInvalidBirthdayDate          ErrorLocale = "INVALID_BIRTHDAY_DATE"
	LocaleNotFound                     ErrorLocale = "NOT_FOUND"
	LocaleUndefined                    ErrorLocale = "UNDEFINED"
	LocaleAuthEmailNotFound            ErrorLocale = "AUTH_EMAIL_NOT_FOUND"
	LocaleAuthEmailAlreadyRegistered   ErrorLocale = "AUTH_EMAIL_ALREADY_REGISTERED"
	LocaleAuthResetTokenNotFound       ErrorLocale = "AUTH_RESET_TOKEN_NOT_FOUND"
	LocaleAuthResetTokenExpired        ErrorLocale = "AUTH_RESET_TOKEN_EXPIRED"
	LocaleAuthPasswordNotMatch         ErrorLocale = "AUTH_PASSWORD_NOT_MATCH"
	LocaleAuthConfirmationCodeNotFound ErrorLocale = "AUTH_CONFIRMATION_CODE_NOT_FOUND"
	LocaleAuthConfirmationCodeExpired  ErrorLocale = "AUTH_CONFIRMATION_CODE_EXPIRED"
	LocaleUserNotVerified              ErrorLocale = "USER_NOT_VERIFIED"
)

type ErrorStatus uint

const (
	StatusUndefined ErrorStatus = iota
	StatusUnauthorized
	StatusBadRequest
	StatusForbidden
	StatusConflict
	StatusNotFound
	StatusInternalServer
	StatusNotAcceptable
)

type CustomError struct {
	status   ErrorStatus
	locale   ErrorLocale
	original error
}

func (status ErrorStatus) New(msg string) error {
	return CustomError{status: status, original: m_errors.New(msg)}
}

func (status ErrorStatus) LocaleNew(msg string, locale ErrorLocale) error {
	return CustomError{status: status, original: m_errors.New(msg), locale: locale}
}

func (status ErrorStatus) Newf(msg string, args ...interface{}) error {
	return CustomError{status: status, original: fmt.Errorf(msg, args...)}
}

func (status ErrorStatus) LocaleNewf(msg string, locale ErrorLocale, args ...interface{}) error {
	return CustomError{status: status, original: fmt.Errorf(msg, args...), locale: locale}
}

func (status ErrorStatus) Wrap(err error, msg string) error {
	return status.Wrapf(err, msg)
}

func (status ErrorStatus) Wrapf(err error, msg string, args ...interface{}) error {
	return CustomError{status: status, original: m_errors.Wrapf(err, msg, args...)}
}

func (status ErrorStatus) LocaleWrapf(err error, msg string, locale ErrorLocale, args ...interface{}) error {
	if err == nil {
		return CustomError{status: status, original: m_errors.New(msg), locale: locale}
	}
	return CustomError{status: status, original: m_errors.Wrapf(err, msg, args...), locale: locale}
}

func (error CustomError) Error() string {
	if error.original != nil {
		return error.original.Error()
	}

	return "unexpected error"
}

func New(msg string) error {
	return CustomError{status: StatusUndefined, original: m_errors.New(msg)}
}

func LocaleNew(msg string, locale ErrorLocale) error {
	return CustomError{status: StatusUndefined, original: m_errors.New(msg), locale: locale}
}

func Newf(msg string, args ...interface{}) error {
	return CustomError{status: StatusUndefined, original: m_errors.New(fmt.Sprintf(msg, args...))}
}

func LocaleNewf(msg string, locale ErrorLocale, args ...interface{}) error {
	return CustomError{status: StatusUndefined, original: m_errors.New(fmt.Sprintf(msg, args...)), locale: locale}
}

func Wrap(err error, msg string) error {
	return Wrapf(err, msg)
}

func LocaleWrap(err error, msg string, locale ErrorLocale) error {
	return LocaleWrapf(err, msg, locale)
}

func Cause(err error) error {
	return m_errors.Cause(err)
}

func Is(err, target error) bool {
	return m_errors.Is(err, target)
}

func Wrapf(err error, msg string, args ...interface{}) error {
	wrappedError := m_errors.Wrapf(err, msg, args...)
	if customErr, ok := err.(CustomError); ok {
		return CustomError{
			status:   customErr.status,
			original: wrappedError,
		}
	}

	return CustomError{status: StatusUndefined, original: wrappedError}
}

func LocaleWrapf(err error, msg string, locale ErrorLocale, args ...interface{}) error {
	wrappedError := m_errors.Wrapf(err, msg, args...)
	if customErr, ok := err.(CustomError); ok {
		return CustomError{
			status:   customErr.status,
			original: wrappedError,
			locale:   locale,
		}
	}

	return CustomError{status: StatusUndefined, locale: locale, original: wrappedError}
}

func AddLocale(err error, locale ErrorLocale) error {
	if customErr, ok := err.(CustomError); ok {
		return CustomError{status: customErr.status, original: customErr.original, locale: locale}
	}

	return CustomError{status: StatusUndefined, original: err, locale: locale}
}

func GetLocale(err error) ErrorLocale {
	if customErr, ok := err.(CustomError); ok && customErr.locale != "" {
		return customErr.locale
	}

	return LocaleUndefined
}

func GetStatusCode(err error) int {
	if customErr, ok := err.(CustomError); ok {
		switch customErr.status {
		case StatusUnauthorized:
			return http.StatusUnauthorized
		case StatusBadRequest:
			return http.StatusBadRequest
		case StatusForbidden:
			return http.StatusForbidden
		case StatusConflict:
			return http.StatusConflict
		case StatusNotFound:
			return http.StatusNotFound
		case StatusNotAcceptable:
			return http.StatusNotAcceptable
		default:
			return http.StatusInternalServerError
		}
	}

	return http.StatusInternalServerError
}
