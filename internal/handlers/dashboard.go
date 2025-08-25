package handlers

import (
	"net/http"

	"github.com/alpemreelmas/sysara/internal/models"
	templ "github.com/alpemreelmas/sysara/templ"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardHandler handles dashboard operations
type DashboardHandler struct {
	db *gorm.DB
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// ShowDashboard displays the main dashboard
func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	// Get some basic statistics
	var userCount, sshKeyCount, serverCount int64
	h.db.Model(&models.User{}).Count(&userCount)
	h.db.Model(&models.SSHKey{}).Count(&sshKeyCount)
	h.db.Model(&models.Server{}).Count(&serverCount)

	data := templ.DashboardData{
		AuthData: templ.AuthData{
			Title:       "Dashboard - Sysara",
			PageTitle:   "Dashboard",
			CurrentUser: *userModel,
		},
		Stats: templ.DashboardStats{
			Users:   int(userCount),
			SSHKeys: int(sshKeyCount),
			Servers: int(serverCount),
		},
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.Dashboard(data).Render(c.Request.Context(), c.Writer)
}
