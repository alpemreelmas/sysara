package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/alpemreelmas/sysara/internal/middleware"
	"github.com/alpemreelmas/sysara/internal/models"
	templ "github.com/alpemreelmas/sysara/templ"
	"github.com/gin-gonic/gin"
)

// EnvHandler handles environment file operations
type EnvHandler struct{}

// NewEnvHandler creates a new environment handler
func NewEnvHandler() *EnvHandler {
	return &EnvHandler{}
}

// ShowEnvFiles displays available environment files
func (h *EnvHandler) ShowEnvFiles(c *gin.Context) {
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	// Get flash messages from context
	flashMessages := []templ.FlashMessage{}
	if messages, exists := c.Get("flash_messages"); exists {
		if msgs, ok := messages.([]middleware.FlashMessage); ok {
			for _, msg := range msgs {
				flashMessages = append(flashMessages, templ.FlashMessage{
					Type:    msg.Type,
					Message: msg.Message,
				})
			}
		}
	}

	// Get list of .env files in the current directory
	envFiles := []string{}
	files, err := os.ReadDir(".")
	if err == nil {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".env") {
				envFiles = append(envFiles, file.Name())
			}
		}
	}

	data := templ.EnvListData{
		AuthData: templ.AuthData{
			Title:       "Environment Files - Sysara",
			PageTitle:   "Environment Files",
			CurrentUser: *userModel,
		},
		EnvFiles:      envFiles,
		FlashMessages: flashMessages,
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.EnvList(data).Render(c.Request.Context(), c.Writer)
}

// ShowEditEnv displays the environment file editor
func (h *EnvHandler) ShowEditEnv(c *gin.Context) {
	filename := c.Param("filename")
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	// Get flash messages from context
	flashMessages := []templ.FlashMessage{}
	if messages, exists := c.Get("flash_messages"); exists {
		if msgs, ok := messages.([]middleware.FlashMessage); ok {
			for _, msg := range msgs {
				flashMessages = append(flashMessages, templ.FlashMessage{
					Type:    msg.Type,
					Message: msg.Message,
				})
			}
		}
	}

	// Validate filename to prevent directory traversal
	if !strings.HasPrefix(filename, ".env") {
		data := templ.EnvEditData{
			AuthData: templ.AuthData{
				Title:       "Edit Environment - Sysara",
				PageTitle:   "Edit Environment",
				CurrentUser: *userModel,
			},
			Filename:      filename,
			Error:         "Invalid environment file name",
			FlashMessages: flashMessages,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
		return
	}

	// Read file content
	content := ""
	if _, err := os.Stat(filename); err == nil {
		contentBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			data := templ.EnvEditData{
				AuthData: templ.AuthData{
					Title:       "Edit Environment - Sysara",
					PageTitle:   "Edit Environment",
					CurrentUser: *userModel,
				},
				Filename:      filename,
				Error:         "Failed to read environment file",
				FlashMessages: flashMessages,
			}
			c.Header("Content-Type", "text/html")
			c.Status(http.StatusInternalServerError)
			templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
			return
		}
		content = string(contentBytes)
	}

	data := templ.EnvEditData{
		AuthData: templ.AuthData{
			Title:       fmt.Sprintf("Edit %s - Sysara", filename),
			PageTitle:   "Edit Environment",
			CurrentUser: *userModel,
		},
		Filename:      filename,
		Content:       content,
		FlashMessages: flashMessages,
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
}

// UpdateEnv saves changes to an environment file
func (h *EnvHandler) UpdateEnv(c *gin.Context) {
	filename := c.Param("filename")
	content := c.PostForm("content")

	// Validate filename to prevent directory traversal
	if !strings.HasPrefix(filename, ".env") {
		middleware.SetFlash(c, "error", "Invalid environment file name")
		c.Redirect(http.StatusSeeOther, "/env")
		return
	}

	// Create backup of existing file
	if _, err := os.Stat(filename); err == nil {
		backupName := fmt.Sprintf("%s.backup.%d", filename, os.Getpid())
		if err := copyFile(filename, backupName); err != nil {
			middleware.SetFlash(c, "error", "Failed to create backup")
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/env/edit/%s", filename))
			return
		}
	}

	// Write new content
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		middleware.SetFlash(c, "error", "Failed to save environment file")
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/env/edit/%s", filename))
		return
	}

	// Success message
	middleware.SetFlash(c, "success", fmt.Sprintf("Environment file '%s' saved successfully", filename))
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/env/edit/%s", filename))
}

// CreateEnvFile creates a new environment file
func (h *EnvHandler) CreateEnvFile(c *gin.Context) {
	filename := c.PostForm("filename")
	fmt.Printf("CreateEnvFile called with filename: '%s'\n", filename)

	// Validate filename
	if filename == "" {
		middleware.SetFlash(c, "error", "Filename cannot be empty")
		c.Redirect(http.StatusSeeOther, "/env")
		return
	}

	if !strings.HasPrefix(filename, ".env") {
		middleware.SetFlash(c, "error", "Environment file must start with .env")
		c.Redirect(http.StatusSeeOther, "/env")
		return
	}

	// Additional validation for filename
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		middleware.SetFlash(c, "error", "Invalid filename: cannot contain path separators")
		c.Redirect(http.StatusSeeOther, "/env")
		return
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		middleware.SetFlash(c, "warning", fmt.Sprintf("File '%s' already exists", filename))
		c.Redirect(http.StatusSeeOther, "/env")
		return
	}

	if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
		middleware.SetFlash(c, "error", "Failed to create file: "+err.Error())
		c.Redirect(http.StatusSeeOther, "/env")
		return
	}

	// Success message
	middleware.SetFlash(c, "success", fmt.Sprintf("Environment file '%s' created successfully", filename))
	c.Redirect(http.StatusSeeOther, "/env")
}

// Helper function to copy files
func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}
