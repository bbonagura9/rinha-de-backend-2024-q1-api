package db

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/utils"
)

func ConnectDB() *gorm.DB {
	var db *gorm.DB
	var err error

	time.Sleep(5 * time.Second)
	remainingTries := 5
	for remainingTries > 0 {
		dsn := utils.GetEnv("DB_DSN", "")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			PrepareStmt: true,
		})
		if err != nil {
			time.Sleep(5 * time.Second)
			remainingTries--
		} else {
			break
		}
	}

	if err != nil {
		panic("Failed to connect to database")
	}

	sqlDb, err := db.DB()
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetMaxOpenConns(50)

	return db
}
