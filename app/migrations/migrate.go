package migrations

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string    `gorm:"type:varchar(100);unique_index;not null"`
	Password    string    `gorm:"type:varchar(100);not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	DateOfBirth time.Time `gorm:"type:date"`
}

func getDB() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
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
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	return db, nil
}

func Migrate() error {
	db, err := getDB()
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	fmt.Println("Migration completed successfully")
	return nil
}

func ResetAndMigrate() error {
	db, err := getDB()
	if err != nil {
		return err
	}

	// Drop existing tables
	err = db.Migrator().DropTable(&User{})
	if err != nil {
		return fmt.Errorf("failed to drop tables: %v", err)
	}

	// Recreate tables
	err = db.AutoMigrate(&User{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	fmt.Println("Database reset and migration completed successfully")
	return nil
}

func DropTables() error {
	db, err := getDB()
	if err != nil {
		return err
	}

	// Drop existing tables
	err = db.Migrator().DropTable(&User{})
	if err != nil {
		return fmt.Errorf("failed to drop tables: %v", err)
	}

	fmt.Println("Tables dropped successfully")
	return nil
}