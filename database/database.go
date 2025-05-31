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
		// Hanya load .env di dev, Railway CLI inject vars otomatis
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, relying on environment")
		}

		// Baca var MySQL dari Railway
		user := os.Getenv("MYSQLUSER")
		password := os.Getenv("MYSQLPASSWORD")
		host := os.Getenv("MYSQLHOST")
		// port := os.Getenv("MYSQLPORT")
		name := os.Getenv("MYSQLDATABASE")

		// Bangun DSN dengan format yang benar
		// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		//     user, password, host, port, name,
		// )

		var err error
		// db, err = sql.Open("mysql", dsn)
		// if err != nil {
		//     log.Fatalf("sql.Open error: %v", err)
		// }

		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
			user,
			password,
			host,
			name))
		if err != nil {
			log.Fatal(err)
		}
		// Optional: set pool size
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Inisialisasi GORM
		DB, err = gorm.Open(mysql.New(mysql.Config{
			Conn: db,
		}), &gorm.Config{})
		if err != nil {
			log.Fatalf("gorm.Open error: %v", err)
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
