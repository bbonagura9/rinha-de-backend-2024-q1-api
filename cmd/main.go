package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/db"
	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/models"
	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/routes"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			panic("Error loading .env")
		}
	}

  db := db.ConnectDB()
  db.AutoMigrate(&models.Transacao{})
	db.AutoMigrate(&models.Cliente{})
	models.InitClientes(db)

  r := gin.Default()

  r.POST("/clientes/:id/transacoes", routes.PostTrancacoes(db))
  r.Run(":8080")
}
