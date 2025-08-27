package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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

	// Add common env files if they don't exist
	//commonEnvFiles := []string{".env", ".env.production", ".env.testing", ".env.development"}
	//for _, envFile := range commonEnvFiles {
	//	found := false
	//	for _, existing := range envFiles {
	//		if existing == envFile {
	//			found = true
	//			break
	//		}
	//	}
	//	if !found {
	//		envFiles = append(envFiles, envFile)
	//	}
	//}

	data := templ.EnvListData{
		AuthData: templ.AuthData{
			Title:       "Environment Files - Sysara",
			PageTitle:   "Environment Files",
			CurrentUser: *userModel,
		},
		EnvFiles: envFiles,
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

	// Validate filename to prevent directory traversal
	if !strings.HasPrefix(filename, ".env") {
		data := templ.EnvEditData{
			AuthData: templ.AuthData{
				Title:       "Edit Environment - Sysara",
				PageTitle:   "Edit Environment",
				CurrentUser: *userModel,
			},
			Filename: filename,
			Error:    "Invalid environment file name",
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
				Filename: filename,
				Error:    "Failed to read environment file",
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
		Filename: filename,
		Content:  content,
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
}

// UpdateEnv saves changes to an environment file
func (h *EnvHandler) UpdateEnv(c *gin.Context) {
	filename := c.Param("filename")
	content := c.PostForm("content")
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	// Validate filename to prevent directory traversal
	if !strings.HasPrefix(filename, ".env") {
		data := templ.EnvEditData{
			AuthData: templ.AuthData{
				Title:       "Edit Environment - Sysara",
				PageTitle:   "Edit Environment",
				CurrentUser: *userModel,
			},
			Filename: filename,
			Content:  content,
			Error:    "Invalid environment file name",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
		return
	}

	// Create backup of existing file
	if _, err := os.Stat(filename); err == nil {
		backupName := fmt.Sprintf("%s.backup.%d", filename, os.Getpid())
		if err := copyFile(filename, backupName); err != nil {
			data := templ.EnvEditData{
				AuthData: templ.AuthData{
					Title:       fmt.Sprintf("Edit %s - Sysara", filename),
					PageTitle:   "Edit Environment",
					CurrentUser: *userModel,
				},
				Filename: filename,
				Content:  content,
				Error:    "Failed to create backup",
			}
			c.Header("Content-Type", "text/html")
			c.Status(http.StatusInternalServerError)
			templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
			return
		}
	}

	// Write new content
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		data := templ.EnvEditData{
			AuthData: templ.AuthData{
				Title:       fmt.Sprintf("Edit %s - Sysara", filename),
				PageTitle:   "Edit Environment",
				CurrentUser: *userModel,
			},
			Filename: filename,
			Content:  content,
			Error:    "Failed to save environment file",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusInternalServerError)
		templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
		return
	}

	data := templ.EnvEditData{
		AuthData: templ.AuthData{
			Title:       fmt.Sprintf("Edit %s - Sysara", filename),
			PageTitle:   "Edit Environment",
			CurrentUser: *userModel,
		},
		Filename: filename,
		Content:  content,
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.EnvEdit(data).Render(c.Request.Context(), c.Writer)
}

// CreateEnvFile creates a new environment file
func (h *EnvHandler) CreateEnvFile(c *gin.Context) {
	filename := c.PostForm("filename")

	// Validate filename
	if !strings.HasPrefix(filename, ".env") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment file must start with .env"})
		return
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File already exists"})
		return
	}

	// Create empty file
	if err := ioutil.WriteFile(filename, []byte(""), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}

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
