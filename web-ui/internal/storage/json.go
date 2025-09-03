package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"web-ui/internal/models"
)

// JSONUserStorage implements UserStorage interface using JSON file
type JSONUserStorage struct {
	filePath string
	mutex    sync.RWMutex
	users    map[string]*models.User // key is user ID
}

// NewJSONUserStorage creates a new JSON user storage
func NewJSONUserStorage(filePath string) *JSONUserStorage {
	return &JSONUserStorage{
		filePath: filePath,
		users:    make(map[string]*models.User),
	}
}

// Connect loads users from JSON file
func (j *JSONUserStorage) Connect() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	// Create file if it doesn't exist
	if _, err := os.Stat(j.filePath); os.IsNotExist(err) {
		// Create empty JSON file
		emptyData := make(map[string]*models.User)
		return j.saveToFile(emptyData)
	}

	// Read existing file
	data, err := os.ReadFile(j.filePath)
	if err != nil {
		return fmt.Errorf("failed to read users file: %w", err)
	}

	if len(data) == 0 {
		j.users = make(map[string]*models.User)
		return nil
	}

	if err := json.Unmarshal(data, &j.users); err != nil {
		return fmt.Errorf("failed to unmarshal users data: %w", err)
	}

	return nil
}

// Close saves data and cleans up
func (j *JSONUserStorage) Close() error {
	return j.saveToFile(j.users)
}

// IsConnected checks if storage is initialized
func (j *JSONUserStorage) IsConnected() bool {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	return j.users != nil
}

// CreateUser creates a new user
func (j *JSONUserStorage) CreateUser(user *models.User) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	// Check if user already exists
	for _, existingUser := range j.users {
		if existingUser.Username == user.Username {
			return errors.New("username already exists")
		}
		if existingUser.Email == user.Email {
			return errors.New("email already exists")
		}
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true

	// Store user
	j.users[user.ID] = user

	return j.saveToFile(j.users)
}

// GetUserByID retrieves a user by ID
func (j *JSONUserStorage) GetUserByID(id string) (*models.User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	user, exists := j.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// GetUserByUsername retrieves a user by username
func (j *JSONUserStorage) GetUserByUsername(username string) (*models.User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	for _, user := range j.users {
		if user.Username == username {
			// Return a copy to prevent external modification
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, errors.New("user not found")
}

// GetUserByEmail retrieves a user by email
func (j *JSONUserStorage) GetUserByEmail(email string) (*models.User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	for _, user := range j.users {
		if user.Email == email {
			// Return a copy to prevent external modification
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, errors.New("user not found")
}

// UpdateUser updates an existing user
func (j *JSONUserStorage) UpdateUser(user *models.User) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if _, exists := j.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Store updated user
	j.users[user.ID] = user

	return j.saveToFile(j.users)
}

// DeleteUser deletes a user by ID
func (j *JSONUserStorage) DeleteUser(id string) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if _, exists := j.users[id]; !exists {
		return errors.New("user not found")
	}

	delete(j.users, id)

	return j.saveToFile(j.users)
}

// UserExists checks if a user exists by username or email
func (j *JSONUserStorage) UserExists(username, email string) (bool, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	for _, user := range j.users {
		if user.Username == username || user.Email == email {
			return true, nil
		}
	}

	return false, nil
}

// ListUsers returns all users
func (j *JSONUserStorage) ListUsers() ([]*models.User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	users := make([]*models.User, 0, len(j.users))
	for _, user := range j.users {
		// Return copies to prevent external modification
		userCopy := *user
		users = append(users, &userCopy)
	}

	return users, nil
}

// saveToFile saves users data to JSON file
func (j *JSONUserStorage) saveToFile(users map[string]*models.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users data: %w", err)
	}

	if err := os.WriteFile(j.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write users file: %w", err)
	}

	return nil
}
