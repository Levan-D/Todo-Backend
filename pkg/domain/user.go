package domain

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID                  uuid.UUID  `gorm:"primary_key;AUTO_INCREMENT;column:id;type:UUID;default:uuid_generate_v4();" json:"id"`
	Avatar              *string    `gorm:"column:avatar;type:TEXT;" json:"avatar"`
	Email               string     `gorm:"column:email;type:VARCHAR;size:255;" json:"email"`
	Password            string     `gorm:"column:password;type:VARCHAR;size:255;" json:"password"`
	FirstName           string     `gorm:"column:first_name;type:VARCHAR;size:120;" json:"first_name"`
	LastName            string     `gorm:"column:last_name;type:VARCHAR;size:160;" json:"last_name"`
	IsVerified          *bool      `gorm:"column:is_verified;type:BOOL;default:false;" json:"is_verified"`
	ResetPasswordToken  string     `gorm:"column:reset_password_token;type:VARCHAR;size:255;" json:"reset_password_token"`
	ResetPasswordExpire *time.Time `gorm:"column:reset_password_expire;type:TIMESTAMPTZ;" json:"reset_password_expire"`
	CreatedAt           *time.Time `gorm:"column:created_at;type:TIMESTAMPTZ;" json:"created_at"`
	UpdatedAt           *time.Time `gorm:"column:updated_at;type:TIMESTAMPTZ;" json:"updated_at"`
	VerifiedAt          *time.Time `gorm:"column:verified_at;type:TIMESTAMPTZ;" json:"verified_at"`
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) BeforeSave(db *gorm.DB) error {
	return nil
}

func (u *User) Prepare() {
}
