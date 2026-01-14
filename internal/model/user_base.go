package model

import (
	"AI_Chat/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type UserBase struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id,string" redis:"id"`
	Username  string    `gorm:"size:50;uniqueIndex;not null" json:"username" redis:"username"`
	Password  string    `gorm:"size:255;not null" json:"-" redis:"password"`
	Email     string    `gorm:"size:100;uniqueIndex" json:"email" redis:"email"`
	CreatedAt time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt time.Time `json:"updated_at" redis:"updated_at"`
}

func (u *UserBase) TableName() string {
	return "user_base"
}
func (u *UserBase) BeforeCreate(tx *gorm.DB) (err error) {
	//防止其他地方ID被手动设置
	if u.ID == 0 {
		u.ID = utils.GenerateID()
	}
	return nil
}
