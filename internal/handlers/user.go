package handlers

import (
	"net/http"
	"strconv"

	"github.com/alpemreelmas/sysara/internal/auth"
	"github.com/alpemreelmas/sysara/internal/models"
	templ "github.com/alpemreelmas/sysara/templ"
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

	data := templ.LoginData{
		Title: "Login - Sysara",
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.Login(data).Render(c.Request.Context(), c.Writer)
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	user, err := h.authService.AuthenticateUser(email, password)
	if err != nil {
		data := templ.LoginData{
			Title: "Login - Sysara",
			Error: err.Error(),
			Email: email,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.Login(data).Render(c.Request.Context(), c.Writer)
		return
	}

	if err := h.authService.Login(c, user); err != nil {
		data := templ.LoginData{
			Title: "Login - Sysara",
			Error: "Failed to create session",
			Email: email,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusInternalServerError)
		templ.Login(data).Render(c.Request.Context(), c.Writer)
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

	data := templ.RegisterData{
		Title: "Register - Sysara",
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.Register(data).Render(c.Request.Context(), c.Writer)
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	email := c.PostForm("email")
	name := c.PostForm("name")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Validate passwords match
	if password != confirmPassword {
		data := templ.RegisterData{
			Title: "Register - Sysara",
			Error: "Passwords do not match",
			Email: email,
			Name:  name,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.Register(data).Render(c.Request.Context(), c.Writer)
		return
	}

	// Validate password length
	if len(password) < 6 {
		data := templ.RegisterData{
			Title: "Register - Sysara",
			Error: "Password must be at least 6 characters long",
			Email: email,
			Name:  name,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.Register(data).Render(c.Request.Context(), c.Writer)
		return
	}

	user, err := h.authService.RegisterUser(email, name, password)
	if err != nil {
		data := templ.RegisterData{
			Title: "Register - Sysara",
			Error: err.Error(),
			Email: email,
			Name:  name,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.Register(data).Render(c.Request.Context(), c.Writer)
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
		currentUser, _ := c.Get("current_user")
		userModel, ok := currentUser.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
			return
		}

		data := templ.UserListData{
			AuthData: templ.AuthData{
				Title:       "Users - Sysara",
				PageTitle:   "Users",
				CurrentUser: *userModel,
			},
			Users: users,
			Error: "Failed to fetch users",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusInternalServerError)
		templ.UserList(data).Render(c.Request.Context(), c.Writer)
		return
	}

	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	data := templ.UserListData{
		AuthData: templ.AuthData{
			Title:       "Users - Sysara",
			PageTitle:   "Users",
			CurrentUser: *userModel,
		},
		Users: users,
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.UserList(data).Render(c.Request.Context(), c.Writer)
}

// ShowCreateUser displays the create user form
func (h *UserHandler) ShowCreateUser(c *gin.Context) {
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	data := templ.UserCreateData{
		AuthData: templ.AuthData{
			Title:       "Create User - Sysara",
			PageTitle:   "Create User",
			CurrentUser: *userModel,
		},
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.UserCreate(data).Render(c.Request.Context(), c.Writer)
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(c *gin.Context) {
	email := c.PostForm("email")
	name := c.PostForm("name")
	password := c.PostForm("password")

	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	// Validate password length
	if len(password) < 6 {
		data := templ.UserCreateData{
			AuthData: templ.AuthData{
				Title:       "Create User - Sysara",
				PageTitle:   "Create User",
				CurrentUser: *userModel,
			},
			Error: "Password must be at least 6 characters long",
			Email: email,
			Name:  name,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.UserCreate(data).Render(c.Request.Context(), c.Writer)
		return
	}

	_, err := h.authService.RegisterUser(email, name, password)
	if err != nil {
		data := templ.UserCreateData{
			AuthData: templ.AuthData{
				Title:       "Create User - Sysara",
				PageTitle:   "Create User",
				CurrentUser: *userModel,
			},
			Error: err.Error(),
			Email: email,
			Name:  name,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.UserCreate(data).Render(c.Request.Context(), c.Writer)
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
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	data := templ.UserEditData{
		AuthData: templ.AuthData{
			Title:       "Edit User - Sysara",
			PageTitle:   "Edit User",
			CurrentUser: *userModel,
		},
		User: user,
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.UserEdit(data).Render(c.Request.Context(), c.Writer)
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
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	// Update password if provided
	if password != "" {
		if len(password) < 6 {
			data := templ.UserEditData{
				AuthData: templ.AuthData{
					Title:       "Edit User - Sysara",
					PageTitle:   "Edit User",
					CurrentUser: *userModel,
				},
				User:  user,
				Error: "Password must be at least 6 characters long",
			}
			c.Header("Content-Type", "text/html")
			c.Status(http.StatusBadRequest)
			templ.UserEdit(data).Render(c.Request.Context(), c.Writer)
			return
		}

		hashedPassword, err := h.authService.HashPassword(password)
		if err != nil {
			data := templ.UserEditData{
				AuthData: templ.AuthData{
					Title:       "Edit User - Sysara",
					PageTitle:   "Edit User",
					CurrentUser: *userModel,
				},
				User:  user,
				Error: "Failed to hash password",
			}
			c.Header("Content-Type", "text/html")
			c.Status(http.StatusInternalServerError)
			templ.UserEdit(data).Render(c.Request.Context(), c.Writer)
			return
		}
		user.Password = hashedPassword
	}

	// Update fields
	user.Email = email
	user.Name = name

	if err := h.db.Save(&user).Error; err != nil {
		data := templ.UserEditData{
			AuthData: templ.AuthData{
				Title:       "Edit User - Sysara",
				PageTitle:   "Edit User",
				CurrentUser: *userModel,
			},
			User:  user,
			Error: "Failed to update user",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusInternalServerError)
		templ.UserEdit(data).Render(c.Request.Context(), c.Writer)
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
