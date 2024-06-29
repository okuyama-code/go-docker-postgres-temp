package main

import (
	"strconv"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"go-api/migrations"

)

type User struct {
	gorm.Model
	Name         string `gorm:"type:varchar(100);not null"`
	Email        string `gorm:"type:varchar(100);unique_index;not null"`
	Password     string `gorm:"type:varchar(100)"`
	Icon         string `gorm:"type:text"`
	AuthProvider string `gorm:"type:varchar(20);default:'local'"`
}

var DB *gorm.DB

func main() {
	migrateFlag := flag.Bool("migrate", false, "Run database migrations")
	resetFlag := flag.Bool("reset", false, "Reset database and run migrations")
	dropFlag := flag.Bool("drop", false, "Drop all tables")
	flag.Parse()

	if *dropFlag {
		err := migrations.DropTables()
		if err != nil {
			log.Fatalf("Failed to drop tables: %v", err)
		}
		return
	}

	if *resetFlag {
		err := migrations.ResetAndMigrate()
		if err != nil {
			log.Fatalf("Failed to reset and migrate database: %v", err)
		}
		return
	}

	if *migrateFlag {
		err := migrations.Migrate()
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		return
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Add this line to check if the table exists
	if err := DB.AutoMigrate(&User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	r := gin.Default()

	r.POST("/api/v1/auth/register", register)
	r.POST("/api/v1/auth/login", login)
	r.GET("/api/v1/users", getAllUsers)

	r.Run()
}

func runMigrations() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func register(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// Check if user already exists
	var existingUser User
	err := DB.Where("email = ?", user.Email).First(&existingUser).Error
	if err == nil {
		// User already exists, return the existing user
		existingUser.Password = "" // Remove password for security
		c.JSON(http.StatusOK, gin.H{
			"user":    existingUser,
			"message": "既存のユーザーです",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing user"})
		return
	}

	// Set auth_provider
	if user.AuthProvider == "" {
		if user.Password == "" {
			user.AuthProvider = "social"
		} else {
			user.AuthProvider = "local"
		}
	}

	// Handle password for local auth
	if user.AuthProvider == "local" && user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while hashing password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	// Handle icon
	if user.Icon != "" {
		// Your existing icon handling code here
	}

	// Create new user
	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while registering user: " + err.Error()})
		return
	}

	user.Password = "" // Remove password for security
	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "新規登録しました",
	})
}
func login(c *gin.Context) {
	var loginInfo struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&loginInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var user User
	if err := DB.Where("email = ?", loginInfo.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "This account uses social login"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginInfo.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// 新しい getAllUsers 関数
func getAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var users []User
	var total int64

	if err := DB.Model(&User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users"})
		return
	}

	if err := DB.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving users"})
		return
	}

	// パスワードを削除
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page": page,
		"page_size": pageSize,
		"total_pages": (int(total) + pageSize - 1) / pageSize,
	})
}