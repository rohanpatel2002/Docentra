package repository

import (
	"ai-document-assistant/internal/models"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user database operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
}

// userRepository implements UserRepository using GORM
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser inserts a user into the DB
func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail finds a user by their email address
func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// GetUserByID finds a user by their primary ID
func (r *userRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}
