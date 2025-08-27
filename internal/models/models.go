package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User represents a user in the system
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email" binding:"required,email"`
	Name      string    `gorm:"not null" json:"name" binding:"required"`
	Password  string    `gorm:"not null" json:"-"` // Hidden from JSON output
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SSHKey represents an SSH key for server access
type SSHKey struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name" binding:"required"`
	PublicKey   string    `gorm:"type:text;not null" json:"public_key" binding:"required"`
	Fingerprint string    `gorm:"not null" json:"fingerprint"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Server represents a monitored server
type Server struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name" binding:"required"`
	Host        string    `gorm:"not null" json:"host" binding:"required"`
	Port        int       `gorm:"default:22" json:"port"`
	Description string    `json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// InitDB initializes the database connection and runs migrations
func InitDB() (*gorm.DB, error) {
	var err error

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open("data/sysara.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schemas
	err = DB.AutoMigrate(&User{}, &SSHKey{}, &Server{})
	if err != nil {
		return nil, err
	}

	// Create default admin user if no users exist
	var userCount int64
	DB.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

		admin := User{
			Email:    "admin@admin",
			Name:     "Administrator",
			Password: string(hashedPassword),
		}

		if err := DB.Create(&admin).Error; err != nil {
			return nil, err
		}
	}

	return DB, nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
