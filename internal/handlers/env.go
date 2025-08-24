package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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

	// Get list of .env files in the current directory
	envFiles := []string{}
	files, err := ioutil.ReadDir(".")
	if err == nil {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".env") {
				envFiles = append(envFiles, file.Name())
			}
		}
	}

	// Add common env files if they don't exist
	commonEnvFiles := []string{".env", ".env.production", ".env.testing", ".env.development"}
	for _, envFile := range commonEnvFiles {
		found := false
		for _, existing := range envFiles {
			if existing == envFile {
				found = true
				break
			}
		}
		if !found {
			envFiles = append(envFiles, envFile)
		}
	}

	c.HTML(http.StatusOK, "pages/env/list.html", gin.H{
		"Title":       "Environment Files - Sysara",
		"CurrentUser": currentUser,
		"EnvFiles":    envFiles,
	})
}

// ShowEditEnv displays the environment file editor
func (h *EnvHandler) ShowEditEnv(c *gin.Context) {
	filename := c.Param("filename")
	currentUser, _ := c.Get("current_user")

	// Validate filename to prevent directory traversal
	if !strings.HasPrefix(filename, ".env") {
		c.HTML(http.StatusBadRequest, "pages/env/edit.html", gin.H{
			"Title":       "Edit Environment - Sysara",
			"CurrentUser": currentUser,
			"Error":       "Invalid environment file name",
			"Filename":    filename,
		})
		return
	}

	// Read file content
	content := ""
	if _, err := os.Stat(filename); err == nil {
		contentBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "pages/env/edit.html", gin.H{
				"Title":       "Edit Environment - Sysara",
				"CurrentUser": currentUser,
				"Error":       "Failed to read environment file",
				"Filename":    filename,
			})
			return
		}
		content = string(contentBytes)
	}

	c.HTML(http.StatusOK, "pages/env/edit.html", gin.H{
		"Title":       fmt.Sprintf("Edit %s - Sysara", filename),
		"CurrentUser": currentUser,
		"Filename":    filename,
		"Content":     content,
	})
}

// UpdateEnv saves changes to an environment file
func (h *EnvHandler) UpdateEnv(c *gin.Context) {
	filename := c.Param("filename")
	content := c.PostForm("content")
	currentUser, _ := c.Get("current_user")

	// Validate filename to prevent directory traversal
	if !strings.HasPrefix(filename, ".env") {
		c.HTML(http.StatusBadRequest, "pages/env/edit.html", gin.H{
			"Title":       "Edit Environment - Sysara",
			"CurrentUser": currentUser,
			"Error":       "Invalid environment file name",
			"Filename":    filename,
			"Content":     content,
		})
		return
	}

	// Create backup of existing file
	if _, err := os.Stat(filename); err == nil {
		backupName := fmt.Sprintf("%s.backup.%d", filename, os.Getpid())
		if err := copyFile(filename, backupName); err != nil {
			c.HTML(http.StatusInternalServerError, "pages/env/edit.html", gin.H{
				"Title":       fmt.Sprintf("Edit %s - Sysara", filename),
				"CurrentUser": currentUser,
				"Error":       "Failed to create backup",
				"Filename":    filename,
				"Content":     content,
			})
			return
		}
	}

	// Write new content
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		c.HTML(http.StatusInternalServerError, "pages/env/edit.html", gin.H{
			"Title":       fmt.Sprintf("Edit %s - Sysara", filename),
			"CurrentUser": currentUser,
			"Error":       "Failed to save environment file",
			"Filename":    filename,
			"Content":     content,
		})
		return
	}

	c.HTML(http.StatusOK, "pages/env/edit.html", gin.H{
		"Title":       fmt.Sprintf("Edit %s - Sysara", filename),
		"CurrentUser": currentUser,
		"Filename":    filename,
		"Content":     content,
		"Success":     "Environment file saved successfully",
	})
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