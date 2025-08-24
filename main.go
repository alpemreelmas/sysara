package main

import (
	"github.com/alpemreelmas/sysara/internal/auth"
	"github.com/alpemreelmas/sysara/internal/handlers"
	"github.com/alpemreelmas/sysara/internal/middleware"
	"github.com/alpemreelmas/sysara/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func main() {
	// Initialize database
	db, err := models.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize session store
	store := sessions.NewCookieStore([]byte("sysara-secret-key-change-in-production"))

	// Initialize auth service
	authService := auth.NewAuthService(db, store)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db, authService)
	dashboardHandler := handlers.NewDashboardHandler(db)
	envHandler := handlers.NewEnvHandler()
	sshHandler := handlers.NewSSHHandler(db)
	monitorHandler := handlers.NewMonitorHandler()

	// Set Gin to release mode in production
	gin.SetMode(gin.DebugMode) // Change to gin.ReleaseMode in production

	// Initialize Gin router
	r := gin.Default()

	// Load HTML templates
	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/**/*.html")
	// Middleware
	r.Use(middleware.SessionMiddleware(store))
	r.Use(middleware.CORSMiddleware())

	// Public routes
	public := r.Group("/")
	{
		public.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/login")
		})
		public.GET("/login", userHandler.ShowLogin)
		public.POST("/login", userHandler.Login)
		public.GET("/register", userHandler.ShowRegister)
		public.POST("/register", userHandler.Register)
	}

	// Protected routes (require authentication)
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// Dashboard
		protected.GET("/dashboard", dashboardHandler.ShowDashboard)

		// User management
		users := protected.Group("/users")
		{
			users.GET("/", userHandler.ListUsers)
			users.GET("/create", userHandler.ShowCreateUser)
			users.POST("/create", userHandler.CreateUser)
			users.GET("/:id/edit", userHandler.ShowEditUser)
			users.POST("/:id/edit", userHandler.UpdateUser)
			users.POST("/:id/delete", userHandler.DeleteUser)
		}

		// Environment management
		env := protected.Group("/env")
		{
			env.GET("/", envHandler.ShowEnvFiles)
			env.GET("/edit/:filename", envHandler.ShowEditEnv)
			env.POST("/edit/:filename", envHandler.UpdateEnv)
			env.POST("/create", envHandler.CreateEnvFile)
		}

		// SSH Key management
		ssh := protected.Group("/ssh")
		{
			ssh.GET("/", sshHandler.ListKeys)
			ssh.GET("/create", sshHandler.ShowCreateKey)
			ssh.POST("/create", sshHandler.CreateKey)
			ssh.POST("/:id/delete", sshHandler.DeleteKey)
		}

		// System monitoring
		monitor := protected.Group("/monitor")
		{
			monitor.GET("/", monitorHandler.ShowMonitor)
			monitor.GET("/api/stats", monitorHandler.GetSystemStats)   // HTMX endpoint
			monitor.GET("/api/processes", monitorHandler.GetProcesses) // HTMX endpoint
		}

		// Logout
		protected.POST("/logout", userHandler.Logout)
	}

	// Start server
	log.Println("Starting Sysara server on :8080")
	log.Fatal(r.Run(":8080"))
}
