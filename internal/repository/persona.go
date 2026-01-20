package repository

import (
	"AI_Chat/internal/model"

	"gorm.io/gorm"
)

type PersonaRepository struct {
	db *gorm.DB
}

func NewPersonaRepository(db *gorm.DB) *PersonaRepository {
	return &PersonaRepository{db: db}
}

func (r *PersonaRepository) CreatePersona(persona *model.Persona) error {
	return r.db.Create(persona).Error
}

func (r *PersonaRepository) GetPersonaById(id string) (*model.Persona, error) {
	var persona model.Persona
	if err := r.db.Where("id = ?", id).First(&persona).Error; err != nil {
		return nil, err
	}
	return &persona, nil
}
func (r *PersonaRepository) GetPersonasByUserId(userId int64) ([]model.Persona, error) {
	var personas []model.Persona
	if err := r.db.Where("user_id = ?", userId).Find(&personas).Error; err != nil {
		return nil, err
	}
	return personas, nil
}
