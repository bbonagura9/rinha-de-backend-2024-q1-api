package db

import (
  "os"
  "time"

  "gorm.io/gorm"
  "gorm.io/driver/postgres"
  "gorm.io/driver/sqlite"

  "github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/utils"
)

func ConnectDB() *gorm.DB {
  var db *gorm.DB
  var err error
  var dialector gorm.Dialector

  dbType := os.Getenv("DB_ENGINE")
  
  remainingTries := 5
  for remainingTries > 0 {
    if dbType == "POSTGRES" {
      dsn := utils.GetEnv("DB_DSN", "")
      dialector = postgres.Open(dsn)
    } else {
      filename := utils.GetEnv("DB_FILE", "todo.db")
      dialector = sqlite.Open(filename)
    }

    db, err = gorm.Open(dialector, &gorm.Config{})
    if err != nil {
      time.Sleep(3 * time.Second)
      remainingTries--
    } else {
      break
    }
  }

  if err != nil {
    panic("Failed to connect to database")
  }

  return db
}
