package models

import (
	"fmt"
	"log"
	"os"

	// Import GORM-related packages.
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//DB 資料庫連接
var DB *gorm.DB

//Setup 資料庫連接設定
func Setup() {
	var err error
	if os.Getenv("GIN_MOD") != "release" {
		err = godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
	}
	if gin.Mode() == "test" {
		DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	} else {
		username := os.Getenv("USERNAME")
		host := os.Getenv("HOST")
		password := os.Getenv("PASSWORD")
		port := os.Getenv("PORT")
		dbName := os.Getenv("DBNAME")
		caRoot := os.Getenv("CAROOT")
		cluster := os.Getenv("CLUSTER")
		var addr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, dbName)
		addr += fmt.Sprintf("?sslmode=verify-full&sslrootcert=%s&options=--cluster=%s", caRoot, cluster)
		DB, err = gorm.Open(postgres.Open(addr), &gorm.Config{})
	}
	if err != nil {
		log.Fatal(err)
	}
	AutoMigrateAll()
}

//AutoMigrateAll 自動產生 table
func AutoMigrateAll() {
	DB.AutoMigrate(&User{})
}
