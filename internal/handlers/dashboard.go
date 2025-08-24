package handlers

import (
	"net/http"

	"github.com/alpemreelmas/sysara/internal/models"
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

	// Get some basic statistics
	var userCount, sshKeyCount, serverCount int64
	h.db.Model(&models.User{}).Count(&userCount)
	h.db.Model(&models.SSHKey{}).Count(&sshKeyCount)
	h.db.Model(&models.Server{}).Count(&serverCount)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Title":       "Dashboard - Sysara",
		"CurrentUser": currentUser,
		"Stats": gin.H{
			"Users":   userCount,
			"SSHKeys": sshKeyCount,
			"Servers": serverCount,
		},
	})
}
