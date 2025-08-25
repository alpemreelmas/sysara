package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/alpemreelmas/sysara/internal/models"
	templ "github.com/alpemreelmas/sysara/templ"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SSHHandler handles SSH key operations
type SSHHandler struct {
	db *gorm.DB
}

// NewSSHHandler creates a new SSH handler
func NewSSHHandler(db *gorm.DB) *SSHHandler {
	return &SSHHandler{db: db}
}

// ListKeys displays all SSH keys
func (h *SSHHandler) ListKeys(c *gin.Context) {
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	var sshKeys []models.SSHKey
	if err := h.db.Preload("User").Find(&sshKeys).Error; err != nil {
		data := templ.SSHListData{
			AuthData: templ.AuthData{
				Title:       "SSH Keys - Sysara",
				PageTitle:   "SSH Keys",
				CurrentUser: *userModel,
			},
			Error: "Failed to fetch SSH keys",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusInternalServerError)
		templ.SSHList(data).Render(c.Request.Context(), c.Writer)
		return
	}

	data := templ.SSHListData{
		AuthData: templ.AuthData{
			Title:       "SSH Keys - Sysara",
			PageTitle:   "SSH Keys",
			CurrentUser: *userModel,
		},
		SSHKeys: sshKeys,
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.SSHList(data).Render(c.Request.Context(), c.Writer)
}

// ShowCreateKey displays the create SSH key form
func (h *SSHHandler) ShowCreateKey(c *gin.Context) {
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	data := templ.SSHCreateData{
		AuthData: templ.AuthData{
			Title:       "Add SSH Key - Sysara",
			PageTitle:   "Add SSH Key",
			CurrentUser: *userModel,
		},
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.SSHCreate(data).Render(c.Request.Context(), c.Writer)
}

// CreateKey handles SSH key creation
func (h *SSHHandler) CreateKey(c *gin.Context) {
	name := c.PostForm("name")
	publicKey := c.PostForm("public_key")
	currentUser, _ := c.Get("current_user")

	user, ok := currentUser.(*models.User)
	if !ok {
		data := templ.SSHCreateData{
			AuthData: templ.AuthData{
				Title:       "Add SSH Key - Sysara",
				PageTitle:   "Add SSH Key",
				CurrentUser: models.User{}, // Empty user as fallback
			},
			Error: "Failed to get current user",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusInternalServerError)
		templ.SSHCreate(data).Render(c.Request.Context(), c.Writer)
		return
	}

	// Validate public key format
	publicKey = strings.TrimSpace(publicKey)
	if !isValidSSHPublicKey(publicKey) {
		data := templ.SSHCreateData{
			AuthData: templ.AuthData{
				Title:       "Add SSH Key - Sysara",
				PageTitle:   "Add SSH Key",
				CurrentUser: *user,
			},
			Name:      name,
			PublicKey: publicKey,
			Error:     "Invalid SSH public key format",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.SSHCreate(data).Render(c.Request.Context(), c.Writer)
		return
	}

	// Generate fingerprint
	fingerprint := generateFingerprint(publicKey)

	// Check if key already exists
	var existingKey models.SSHKey
	if err := h.db.Where("fingerprint = ?", fingerprint).First(&existingKey).Error; err == nil {
		data := templ.SSHCreateData{
			AuthData: templ.AuthData{
				Title:       "Add SSH Key - Sysara",
				PageTitle:   "Add SSH Key",
				CurrentUser: *user,
			},
			Name:      name,
			PublicKey: publicKey,
			Error:     "SSH key already exists",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		templ.SSHCreate(data).Render(c.Request.Context(), c.Writer)
		return
	}

	// Create SSH key
	sshKey := models.SSHKey{
		Name:        name,
		PublicKey:   publicKey,
		Fingerprint: fingerprint,
		UserID:      user.ID,
	}

	if err := h.db.Create(&sshKey).Error; err != nil {
		data := templ.SSHCreateData{
			AuthData: templ.AuthData{
				Title:       "Add SSH Key - Sysara",
				PageTitle:   "Add SSH Key",
				CurrentUser: *user,
			},
			Name:      name,
			PublicKey: publicKey,
			Error:     "Failed to save SSH key",
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusInternalServerError)
		templ.SSHCreate(data).Render(c.Request.Context(), c.Writer)
		return
	}

	c.Redirect(http.StatusSeeOther, "/ssh")
}

// DeleteKey handles SSH key deletion
func (h *SSHHandler) DeleteKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SSH key ID"})
		return
	}

	// Check if key exists and belongs to current user or user is admin
	var sshKey models.SSHKey
	if err := h.db.First(&sshKey, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
		return
	}

	currentUser, exists := c.Get("current_user")
	if exists {
		if user, ok := currentUser.(*models.User); ok {
			// Allow deletion if it's the user's own key
			// In a more complex system, you might have admin roles
			if sshKey.UserID != user.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own SSH keys"})
				return
			}
		}
	}

	if err := h.db.Delete(&sshKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete SSH key"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/ssh")
}

// isValidSSHPublicKey validates SSH public key format
func isValidSSHPublicKey(key string) bool {
	parts := strings.Fields(key)
	if len(parts) < 2 {
		return false
	}

	keyType := parts[0]
	validTypes := []string{"ssh-rsa", "ssh-dss", "ssh-ed25519", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521"}

	for _, validType := range validTypes {
		if keyType == validType {
			return true
		}
	}

	return false
}

// generateFingerprint creates a simple fingerprint for the SSH key
func generateFingerprint(publicKey string) string {
	hash := md5.Sum([]byte(publicKey))
	return fmt.Sprintf("%x", hash)
}
