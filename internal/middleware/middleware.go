package middleware

import (
	"net/http"

	"github.com/alpemreelmas/sysara/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

// SessionMiddleware adds session store to the context
func SessionMiddleware(store sessions.Store) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Store the session store in context for later use
		c.Set("session_store", store)
		c.Next()
	})
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// AuthMiddleware checks if user is authenticated
func AuthMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if !authService.IsAuthenticated(c) {
			// For HTMX requests, return 401 to trigger client-side redirect
			if c.GetHeader("HX-Request") == "true" {
				c.Header("HX-Redirect", "/login")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			// For regular requests, redirect to login
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		
		// Add current user to context
		user, err := authService.GetCurrentUser(c)
		if err != nil {
			if c.GetHeader("HX-Request") == "true" {
				c.Header("HX-Redirect", "/login")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		
		c.Set("current_user", user)
		c.Next()
	})
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})
}