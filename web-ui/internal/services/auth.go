package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"web-ui/internal/models"
	"web-ui/internal/storage"
	"web-ui/internal/utils"
)

// AuthService handles authentication operations
type AuthService struct {
	userStorage storage.UserStorage
	jwtManager  *utils.JWTManager
}

// NewAuthService creates a new authentication service
func NewAuthService(userStorage storage.UserStorage) *AuthService {
	return &AuthService{
		userStorage: userStorage,
		jwtManager:  utils.NewJWTManager(),
	}
}

// SignUp creates a new user account
func (a *AuthService) SignUp(signupData *models.UserSignup) (*models.LoginResponse, error) {
	// Validate input
	if err := a.validateSignupData(signupData); err != nil {
		return nil, err
	}

	// Check if user already exists
	exists, err := a.userStorage.UserExists(signupData.Username, signupData.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, errors.New("user with this username or email already exists")
	}

	// Hash password
	hashedPassword, err := a.hashPassword(signupData.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     strings.TrimSpace(signupData.Username),
		Email:        strings.ToLower(strings.TrimSpace(signupData.Email)),
		PasswordHash: hashedPassword,
	}

	// Save user
	if err := a.userStorage.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, err := a.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return response (don't include password hash)
	userResponse := *user
	userResponse.PasswordHash = ""

	return &models.LoginResponse{
		Token: token,
		User:  userResponse,
	}, nil
}

// Login authenticates a user and returns a token
func (a *AuthService) Login(credentials *models.UserCredentials) (*models.LoginResponse, error) {
	// Validate input
	if err := a.validateLoginData(credentials); err != nil {
		return nil, err
	}

	// Get user by username or email
	var user *models.User
	var err error

	// Check if input is email or username
	if a.isEmail(credentials.Username) {
		user, err = a.userStorage.GetUserByEmail(credentials.Username)
	} else {
		user, err = a.userStorage.GetUserByUsername(credentials.Username)
	}

	if err != nil {
		return nil, errors.New("invalid username/email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if !a.verifyPassword(credentials.Password, user.PasswordHash) {
		return nil, errors.New("invalid username/email or password")
	}

	// Generate token
	token, err := a.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return response (don't include password hash)
	userResponse := *user
	userResponse.PasswordHash = ""

	return &models.LoginResponse{
		Token: token,
		User:  userResponse,
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (a *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	claims, err := a.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get user from storage to ensure they still exist and are active
	user, err := a.userStorage.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Don't return password hash
	user.PasswordHash = ""
	return user, nil
}

// RefreshToken generates a new token from an existing valid token
func (a *AuthService) RefreshToken(tokenString string) (string, error) {
	return a.jwtManager.RefreshToken(tokenString)
}

// ChangePassword changes a user's password
func (a *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	// Get user
	user, err := a.userStorage.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if !a.verifyPassword(currentPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if err := a.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := a.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.PasswordHash = hashedPassword
	if err := a.userStorage.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// hashPassword hashes a plain text password
func (a *AuthService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword verifies a plain text password against a hash
func (a *AuthService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// validateSignupData validates signup input
func (a *AuthService) validateSignupData(data *models.UserSignup) error {
	if strings.TrimSpace(data.Username) == "" {
		return errors.New("username is required")
	}

	if len(strings.TrimSpace(data.Username)) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if len(strings.TrimSpace(data.Username)) > 50 {
		return errors.New("username must be less than 50 characters")
	}

	// Username should only contain alphanumeric characters and underscores
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(strings.TrimSpace(data.Username)) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}

	if strings.TrimSpace(data.Email) == "" {
		return errors.New("email is required")
	}

	if !a.isEmail(data.Email) {
		return errors.New("invalid email format")
	}

	return a.validatePassword(data.Password)
}

// validateLoginData validates login input
func (a *AuthService) validateLoginData(data *models.UserCredentials) error {
	if strings.TrimSpace(data.Username) == "" {
		return errors.New("username/email is required")
	}

	if strings.TrimSpace(data.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}

// validatePassword validates password requirements
func (a *AuthService) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must be less than 128 characters")
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}

// isEmail checks if a string is a valid email format
func (a *AuthService) isEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strings.TrimSpace(email))
}
