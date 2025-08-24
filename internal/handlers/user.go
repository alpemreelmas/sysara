package handlers

import (
	"net/http"
	"strconv"

	"github.com/alpemreelmas/sysara/internal/auth"
	"github.com/alpemreelmas/sysara/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler handles user-related operations
type UserHandler struct {
	db          *gorm.DB
	authService *auth.AuthService
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB, authService *auth.AuthService) *UserHandler {
	return &UserHandler{
		db:          db,
		authService: authService,
	}
}

// ShowLogin displays the login page
func (h *UserHandler) ShowLogin(c *gin.Context) {
	// If already logged in, redirect to dashboard
	if h.authService.IsAuthenticated(c) {
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "Login - Sysara",
	})
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	user, err := h.authService.AuthenticateUser(email, password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title": "Login - Sysara",
			"Error": err.Error(),
			"Email": email,
		})
		return
	}

	if err := h.authService.Login(c, user); err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"Title": "Login - Sysara",
			"Error": "Failed to create session",
			"Email": email,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

// ShowRegister displays the registration page
func (h *UserHandler) ShowRegister(c *gin.Context) {
	// If already logged in, redirect to dashboard
	if h.authService.IsAuthenticated(c) {
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}

	c.HTML(http.StatusOK, "register.html", gin.H{
		"Title": "Register - Sysara",
	})
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	email := c.PostForm("email")
	name := c.PostForm("name")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Validate passwords match
	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register - Sysara",
			"Error": "Passwords do not match",
			"Email": email,
			"Name":  name,
		})
		return
	}

	// Validate password length
	if len(password) < 6 {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register - Sysara",
			"Error": "Password must be at least 6 characters long",
			"Email": email,
			"Name":  name,
		})
		return
	}

	user, err := h.authService.RegisterUser(email, name, password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register - Sysara",
			"Error": err.Error(),
			"Email": email,
			"Name":  name,
		})
		return
	}

	// Automatically log in the user after registration
	if err := h.authService.Login(c, user); err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Logout handles user logout
func (h *UserHandler) Logout(c *gin.Context) {
	if err := h.authService.Logout(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
}

// ListUsers displays all users (admin function)
func (h *UserHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "pages/users/list.html", gin.H{
			"Title": "Users - Sysara",
			"Error": "Failed to fetch users",
		})
		return
	}

	currentUser, _ := c.Get("current_user")

	c.HTML(http.StatusOK, "pages/users/list.html", gin.H{
		"Title":       "Users - Sysara",
		"Users":       users,
		"CurrentUser": currentUser,
	})
}

// ShowCreateUser displays the create user form
func (h *UserHandler) ShowCreateUser(c *gin.Context) {
	currentUser, _ := c.Get("current_user")

	c.HTML(http.StatusOK, "pages/users/create.html", gin.H{
		"Title":       "Create User - Sysara",
		"CurrentUser": currentUser,
	})
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(c *gin.Context) {
	email := c.PostForm("email")
	name := c.PostForm("name")
	password := c.PostForm("password")

	currentUser, _ := c.Get("current_user")

	// Validate password length
	if len(password) < 6 {
		c.HTML(http.StatusBadRequest, "pages/users/create.html", gin.H{
			"Title":       "Create User - Sysara",
			"Error":       "Password must be at least 6 characters long",
			"Email":       email,
			"Name":        name,
			"CurrentUser": currentUser,
		})
		return
	}

	_, err := h.authService.RegisterUser(email, name, password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "pages/users/create.html", gin.H{
			"Title":       "Create User - Sysara",
			"Error":       err.Error(),
			"Email":       email,
			"Name":        name,
			"CurrentUser": currentUser,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/users")
}

// ShowEditUser displays the edit user form
func (h *UserHandler) ShowEditUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/users")
		return
	}

	var user models.User
	if err := h.db.First(&user, uint(id)).Error; err != nil {
		c.Redirect(http.StatusSeeOther, "/users")
		return
	}

	currentUser, _ := c.Get("current_user")

	c.HTML(http.StatusOK, "pages/users/edit.html", gin.H{
		"Title":       "Edit User - Sysara",
		"User":        user,
		"CurrentUser": currentUser,
	})
}

// UpdateUser handles user updates
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/users")
		return
	}

	var user models.User
	if err := h.db.First(&user, uint(id)).Error; err != nil {
		c.Redirect(http.StatusSeeOther, "/users")
		return
	}

	email := c.PostForm("email")
	name := c.PostForm("name")
	password := c.PostForm("password")

	currentUser, _ := c.Get("current_user")

	// Update fields
	user.Email = email
	user.Name = name

	// Update password if provided
	if password != "" {
		if len(password) < 6 {
			c.HTML(http.StatusBadRequest, "pages/users/edit.html", gin.H{
				"Title":       "Edit User - Sysara",
				"Error":       "Password must be at least 6 characters long",
				"User":        user,
				"CurrentUser": currentUser,
			})
			return
		}

		hashedPassword, err := h.authService.HashPassword(password)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "pages/users/edit.html", gin.H{
				"Title":       "Edit User - Sysara",
				"Error":       "Failed to hash password",
				"User":        user,
				"CurrentUser": currentUser,
			})
			return
		}
		user.Password = hashedPassword
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "pages/users/edit.html", gin.H{
			"Title":       "Edit User - Sysara",
			"Error":       "Failed to update user",
			"User":        user,
			"CurrentUser": currentUser,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/users")
}

// DeleteUser handles user deletion
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Don't allow users to delete themselves
	currentUser, exists := c.Get("current_user")
	if exists {
		if user, ok := currentUser.(*models.User); ok && user.ID == uint(id) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
			return
		}
	}

	if err := h.db.Delete(&models.User{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/users")
}
