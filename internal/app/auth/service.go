package auth

import (
	"github.com/Levan-D/Todo-Backend/internal/app/errors"
	"github.com/Levan-D/Todo-Backend/pkg/argon2id"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"strings"
	"time"
)

type service struct {
	repository Repository
	argon      argon2id.Argon2ID
}

type Service interface {
	Login(email string, password string, ip string, agent string) (domain.User, error)
	SingUp(input SignUpInput) (uuid.UUID, error)
	GetByEmail(email string) (domain.User, error)
	ForgotPassword(email string) error
	CheckResetToken(token string) error
	ResetPassword(confirmationCode string, newPassword string, confirmPassword string) error
}

type AuthSocialInput struct {
	Provider     string
	SocialNumber string
	Email        string
	FirstName    string
	LastName     string
	AvatarURL    string
	IP           string
	UserAgent    string
}

type SignUpInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

func NewService(repository Repository, argon argon2id.Argon2ID) Service {
	return &service{
		repository: repository,
		argon:      argon,
	}
}

func (s *service) Login(email string, password string, ip string, agent string) (domain.User, error) {
	user, err := s.repository.FindUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, errors.StatusBadRequest.LocaleWrapf(err, "retrieve a user", errors.LocaleInvalidCredentials)
		}
		return domain.User{}, errors.StatusInternalServer.LocaleWrapf(err, "retrieve a user", errors.LocaleUndefined)
	}

	verify, err := s.argon.Verify(password, user.Password)
	if err != nil {
		return domain.User{}, errors.StatusInternalServer.LocaleWrapf(err, "compare passwords", errors.LocaleUndefined)
	}

	if !verify {
		return domain.User{}, errors.StatusBadRequest.LocaleNew("invalid credentials", errors.LocaleInvalidCredentials)
	}

	return user, nil
}

func (s *service) SingUp(input SignUpInput) (uuid.UUID, error) {
	hashPassword, err := s.argon.Hash(input.Password)
	if err != nil {
		return uuid.UUID{}, errors.StatusInternalServer.LocaleWrapf(err, "cannot be parsed birthday", errors.LocaleUndefined)
	}

	user, err := s.repository.Create(domain.User{
		Email:      input.Email,
		Password:   hashPassword,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		IsVerified: utils.NewFalse(),
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return uuid.UUID{}, errors.StatusConflict.LocaleWrapf(err, "email already registered", errors.LocaleAuthEmailAlreadyRegistered)
		}
		return uuid.UUID{}, errors.StatusInternalServer.LocaleWrapf(err, "cannot be created new user", errors.LocaleUndefined)
	}

	return user.ID, nil
}

func (s *service) GetByEmail(email string) (domain.User, error) {
	item, err := s.repository.FindUserByEmail(email)
	if err != nil {
		return domain.User{}, err
	}
	return item, nil
}

func (s *service) ForgotPassword(email string) error {
	user, err := s.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.StatusNotFound.LocaleWrapf(err, "email not found", errors.LocaleAuthEmailNotFound)
		}
		return errors.StatusInternalServer.LocaleWrapf(err, "email not found", errors.LocaleUndefined)
	}

	// generate reset code
	generateCode := utils.GenerateForgotConfirmationCode(12)

	err = s.repository.UpdateResetData(user.ID, generateCode)
	if err != nil {
		return errors.StatusInternalServer.LocaleWrapf(err, "cannot be update reset data", errors.LocaleUndefined)
	}

	return nil
}

func (s *service) CheckResetToken(token string) error {
	user, err := s.repository.FindByResetToken(token)
	if err != nil {
		return errors.StatusNotFound.LocaleWrapf(err, "reset data not found", errors.LocaleAuthResetTokenNotFound)
	}

	if time.Now().Unix() > user.ResetPasswordExpire.Unix() {
		return errors.StatusNotFound.LocaleNew("reset token has been expired", errors.LocaleAuthResetTokenExpired)
	}

	return nil
}

func (s *service) ResetPassword(confirmationCode string, newPassword string, confirmPassword string) error {
	if newPassword != confirmPassword {
		return errors.StatusBadRequest.LocaleNew("passwords do not match", errors.LocaleAuthPasswordNotMatch)
	}

	hashPassword, err := s.argon.Hash(newPassword)
	if err != nil {
		return errors.StatusBadRequest.LocaleNew("cannot be hash password", errors.LocaleUndefined)
	}

	err = s.repository.UpdatePasswordByConfirmationCode(confirmationCode, hashPassword)
	if err != nil {
		return errors.StatusBadRequest.LocaleNew("cannot be update password by confirmation code", errors.LocaleUndefined)
	}

	err = s.repository.CleanResetData(confirmationCode)
	if err != nil {
		return errors.StatusInternalServer.LocaleNew("cannot be clean reset data", errors.LocaleUndefined)
	}

	return nil
}

func (s *service) CheckConfirmationCode(confirmationCode string) (string, error) {
	user, err := s.repository.FindByForgotConfirmationCode(confirmationCode)
	if err != nil {
		return "", errors.StatusNotFound.LocaleWrapf(err, "cannot be found confirmation code", errors.LocaleAuthConfirmationCodeNotFound)
	}

	if time.Now().Unix() > user.ResetPasswordExpire.Unix() {
		return "", errors.StatusBadRequest.LocaleWrapf(err, "confirmation code has been expired", errors.LocaleAuthConfirmationCodeExpired)
	}

	return user.ResetPasswordToken, nil
}
