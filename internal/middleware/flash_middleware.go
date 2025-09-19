package middleware

import (
	"github.com/alpemreelmas/sysara/pkg/flash"
	"github.com/gin-gonic/gin"
)

// LoadFlashes reads flashes (and clears them) and puts into context for templates
func LoadFlashes() gin.HandlerFunc {
	return func(c *gin.Context) {
		fls := flash.GetAll(c)
		c.Set("flashes", fls)
		c.Next()
	}
}
