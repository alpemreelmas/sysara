package middleware

import (
	"log"
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

// FlashMessage represents a flash message
type FlashMessage struct {
	Type    string `json:"type"`    // success, error, warning, info
	Message string `json:"message"`
}

// FlashMiddleware handles flash messages
func FlashMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get session store from context
		store, exists := c.Get("session_store")
		if !exists {
			c.Next()
			return
		}

		sessionStore := store.(sessions.Store)
		session, err := sessionStore.Get(c.Request, "flash-session")
		if err != nil {
			c.Next()
			return
		}

		// Get flash messages from session
		var flashMessages []FlashMessage
		if flashes := session.Flashes(); len(flashes) > 0 {
			log.Printf("Found %d flash messages", len(flashes))
			for _, flash := range flashes {
				if flashMsg, ok := flash.(FlashMessage); ok {
					flashMessages = append(flashMessages, flashMsg)
					log.Printf("Retrieved flash: type=%s, message=%s", flashMsg.Type, flashMsg.Message)
				}
			}
			// Save session to clear flashes
			session.Save(c.Request, c.Writer)
		}

		// Add flash messages to context
		c.Set("flash_messages", flashMessages)
		log.Printf("Set %d flash messages in context", len(flashMessages))
		c.Next()
	})
}

// SetFlash adds a flash message to the session
func SetFlash(c *gin.Context, msgType, message string) {
	log.Printf("Setting flash message: type=%s, message=%s", msgType, message)
	
	store, exists := c.Get("session_store")
	if !exists {
		log.Println("Session store not found in context")
		return
	}

	sessionStore := store.(sessions.Store)
	session, err := sessionStore.Get(c.Request, "flash-session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return
	}

	flash := FlashMessage{
		Type:    msgType,
		Message: message,
	}

	session.AddFlash(flash)
	err = session.Save(c.Request, c.Writer)
	if err != nil {
		log.Printf("Error saving session: %v", err)
	} else {
		log.Println("Flash message saved successfully")
	}
}