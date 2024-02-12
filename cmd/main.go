package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/db"
	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/routes"
	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/utils"
)

func main() {
	var initClientes bool
	flag.BoolVar(&initClientes, "i", false, "Populate clientes table")

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			panic("Error loading .env")
		}
	}

	ctx := context.Background()

	time.Sleep(5 * time.Second) // Initial wait for db init
	attempts := 5
	pool, err := pgxpool.New(ctx, utils.GetEnv("DB_DSN", ""))
	for err != nil && attempts > 0 {
		pool, err = pgxpool.New(ctx, utils.GetEnv("DB_DSN", ""))
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	queries := db.New(pool)

	r := gin.Default()
	r.POST("/clientes/:id/transacoes", routes.PostTrancacoes(ctx, pool, queries))
	r.GET("/clientes/:id/extrato", routes.GetExtrato(ctx, queries))

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
