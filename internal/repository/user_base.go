package repository

import (
	"AI_Chat/internal/model"

	"gorm.io/gorm"
)

type UserBaseRepository struct {
	db *gorm.DB
}

func NewUserBaseRepository(db *gorm.DB) *UserBaseRepository {
	return &UserBaseRepository{db: db}
}
func (r *UserBaseRepository) CreateUserBase(userBase *model.UserBase) error {
	return r.db.Create(userBase).Error
}
func (r *UserBaseRepository) GetUserBaseByUsername(username string) (*model.UserBase, error) {
	var userBase model.UserBase
	if err := r.db.Where("username = ?", username).Find(&userBase).Error; err != nil {
		return nil, err
	}
	return &userBase, nil
}
