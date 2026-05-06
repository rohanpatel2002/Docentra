package database

import (
	"fmt"
	"log"
	"os"

	"ai-document-assistant/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connecting db
func ConnectDB() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	// Fallback to default if env not set
	if host == "" {
		host = "localhost"
		port = "5432"
		user = "postgres"
		password = "postgres"
		dbname = "aidocdb"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v. Please ensure PostgreSQL is running.", err)
	}
	// Automigrate our schemas
	log.Println("Running database migrations...")
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Failed to migrate database schemas: %v", err)
	}
	log.Println("Successfully connected to the PostgreSQL database")
}
