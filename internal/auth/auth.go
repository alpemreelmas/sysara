package auth

import (
	"errors"
	"net/http"

	"github.com/alpemreelmas/sysara/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication operations
type AuthService struct {
	db    *gorm.DB
	store sessions.Store
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB, store sessions.Store) *AuthService {
	return &AuthService{
		db:    db,
		store: store,
	}
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash
func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// RegisterUser creates a new user with hashed password
func (s *AuthService) RegisterUser(email, name, password string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := models.User{
		Email:    email,
		Name:     name,
		Password: hashedPassword,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// AuthenticateUser verifies user credentials
func (s *AuthService) AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := s.VerifyPassword(user.Password, password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}

// Login creates a user session
func (s *AuthService) Login(c *gin.Context, user *models.User) error {
	session, err := s.store.Get(c.Request, "sysara-session")
	if err != nil {
		return err
	}

	session.Values["user_id"] = user.ID
	session.Values["user_email"] = user.Email
	session.Values["user_name"] = user.Name

	return session.Save(c.Request, c.Writer)
}

// Logout destroys the user session
func (s *AuthService) Logout(c *gin.Context) error {
	session, err := s.store.Get(c.Request, "sysara-session")
	if err != nil {
		return err
	}

	// Clear session values
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1

	return session.Save(c.Request, c.Writer)
}

// GetCurrentUser returns the currently logged-in user
func (s *AuthService) GetCurrentUser(c *gin.Context) (*models.User, error) {
	session, err := s.store.Get(c.Request, "sysara-session")
	if err != nil {
		return nil, err
	}

	userID, ok := session.Values["user_id"].(uint)
	if !ok {
		return nil, errors.New("user not authenticated")
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// IsAuthenticated checks if user is logged in
func (s *AuthService) IsAuthenticated(c *gin.Context) bool {
	session, err := s.store.Get(c.Request, "sysara-session")
	if err != nil {
		return false
	}

	_, ok := session.Values["user_id"].(uint)
	return ok
}

// RequireAuth middleware function
func (s *AuthService) RequireAuth(c *gin.Context) {
	if !s.IsAuthenticated(c) {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}
	c.Next()
}