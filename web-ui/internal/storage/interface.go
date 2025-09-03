package storage

import (
	"web-ui/internal/models"
)

// UserStorage defines the interface for user data storage operations
type UserStorage interface {
	// CreateUser creates a new user in the storage
	CreateUser(user *models.User) error

	// GetUserByID retrieves a user by their ID
	GetUserByID(id string) (*models.User, error)

	// GetUserByUsername retrieves a user by their username
	GetUserByUsername(username string) (*models.User, error)

	// GetUserByEmail retrieves a user by their email
	GetUserByEmail(email string) (*models.User, error)

	// UpdateUser updates an existing user
	UpdateUser(user *models.User) error

	// DeleteUser deletes a user by their ID
	DeleteUser(id string) error

	// UserExists checks if a user exists by username or email
	UserExists(username, email string) (bool, error)

	// ListUsers returns all users (for admin purposes)
	ListUsers() ([]*models.User, error)
}

// DatabaseConnection represents a generic database connection
type DatabaseConnection interface {
	// Connect establishes connection to the database
	Connect() error

	// Close closes the database connection
	Close() error

	// IsConnected checks if the database is connected
	IsConnected() bool
}
