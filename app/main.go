package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"go-api/migrations"
)

type User struct {
	gorm.Model
	Username    string    `gorm:"type:varchar(100);unique_index;not null"`
	Password    string    `gorm:"type:varchar(100);not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	DateOfBirth *time.Time `gorm:"type:date"`
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

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("failed to get database: %v", err)
	}
	defer sqlDB.Close()

	DB.AutoMigrate(&User{})

	r := gin.Default()

	r.POST("/register", register)
	r.POST("/login", login)
	r.GET("/current-user", getCurrentUser)
	r.POST("/logout", logout)

	r.Run()
}

func register(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
	}

	if user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if user.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while hashing password"})
		return
	}
	user.Password = string(hashedPassword)

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while registering user"})
		return
	}

	responseUser := struct {
		ID          uint       `json:"id"`
		Username    string     `json:"username"`
		Name        string     `json:"name"`
	}{
		ID:          user.ID,
		Username:    user.Username,
		Name:        user.Name,
	}

	c.JSON(http.StatusOK, responseUser)
}

func login(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	var foundUser User
	if err := DB.Where("username = ?", user.Username).First(&foundUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT Secret not found"})
		return
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while generating token"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func getCurrentUser(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		var user User
		if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		responseUser := struct {
			ID          uint       `json:"id"`
			Username    string     `json:"username"`
			Name        string     `json:"name"`
			DateOfBirth *time.Time `json:"date_of_birth"`
		}{
			ID:          user.ID,
			Username:    user.Username,
			Name:        user.Name,
			DateOfBirth: user.DateOfBirth,
		}

		c.JSON(http.StatusOK, responseUser)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
	}
}

func logout(c *gin.Context) {
	// サーバーサイドでのログアウト処理は特に必要ありません
	// クライアント側でトークンを削除することでログアウトとみなします
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}