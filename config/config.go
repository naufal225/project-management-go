package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	AppConfig *Config
)

type Config struct {
	AppPort           string
	AppURL            string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	JWTSecret         string
	JWTExpiredMinutes string
	JWTRefreshToken   string
	JWTExpire         string
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found.")
	}

	AppConfig = &Config{
		AppPort:         getEnv("PORT", "8080"),
		AppURL:          getEnv("APP_URL", "http://localhost:8080"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBPassword:      getEnv("DB_PASSWORD", "nma225"),
		DBName:          getEnv("DB_NAME", "project_management"),
		JWTSecret:       getEnv("JWT_SECRET", "KODERAHASIAUNTUKJWTADILIJOKOWIRAJAMULYONOTUKANGTIPU"),
		JWTRefreshToken: getEnv("REFRESH_TOKEN_EXPIRED", "24h"),
		JWTExpire:       getEnv("JWT_EXPIRY", "6h"),
	}

}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	} else {
		return fallback
	}
}

func ConnectDB() {
	cfg := AppConfig

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
}
