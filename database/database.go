package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db     *sql.DB
	dbOnce sync.Once
	DB     *gorm.DB
)

func getDB() *sql.DB {
	dbOnce.Do(func() {
		var err error
		err = godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_URL"),
			os.Getenv("DB_DATABASE")))
		if err != nil {
			log.Fatal(err)
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		err = db.Ping()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		DB, err = gorm.Open(mysql.New(mysql.Config{
			Conn: db,
		}), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}
	})
	return db
}

func GetDBInstance() *sql.DB {
	return getDB()
}

func GetGormDBInstance() *gorm.DB {
	return DB
}
