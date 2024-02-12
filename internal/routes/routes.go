package routes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/db"
)

type PostTransacoesBody struct {
	Valor     int64
	Tipo      string
	Descricao string
}

func PostTrancacoes(ctx context.Context, pool *pgxpool.Pool, q *db.Queries) func(*gin.Context) {
	return func(c *gin.Context) {
		var req PostTransacoesBody
		if err := binding.JSON.Bind(c.Request, &req); err != nil {
			c.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}

		if req.Tipo != "c" && req.Tipo != "d" {
			c.AbortWithError(
				http.StatusUnprocessableEntity,
				errors.New("Invalid: tipo"),
			)
			return
		}

		if req.Descricao == "" || len(req.Descricao) > 10 {
			c.AbortWithError(
				http.StatusUnprocessableEntity,
				errors.New("Invalid: descricao"),
			)
			return
		}

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		fmt.Printf("%+v\n", pool.Stat())
		err = pool.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
			tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				fmt.Println("Failed beginning transaction")
				c.AbortWithError(http.StatusInternalServerError, err)
				return err
			}
			defer func() {
				if err != nil {
					tx.Rollback(ctx)
				} else {
					tx.Commit(ctx)
				}
			}()
			qtx := q.WithTx(tx)

			cliente, err := qtx.GetClienteLock(ctx, id)
			if err != nil {
				if errors.Is(sql.ErrNoRows, err) {
					c.AbortWithStatus(http.StatusNotFound)
				} else {
					c.AbortWithError(http.StatusInternalServerError, err)
				}
				return err
			}

			saldo := cliente.Saldo.Int64
			limite := cliente.Limite.Int64

			if req.Tipo == "d" {
				saldo = saldo - int64(req.Valor)
				if saldo < -limite {
					c.AbortWithStatus(http.StatusUnprocessableEntity)
					return err
				}
			} else if req.Tipo == "c" {
				saldo = saldo + int64(req.Valor)
			}

			err = qtx.UpdateClienteSaldo(
				ctx,
				db.UpdateClienteSaldoParams{
					ID:    id,
					Saldo: pgtype.Int8{Int64: saldo, Valid: true},
				})
			if err != nil {
				fmt.Println("Failed updating cliente saldo")
				c.AbortWithStatus(http.StatusInternalServerError)
				return err
			}

			c.JSON(http.StatusOK, gin.H{
				"limite": cliente.Limite,
				"saldo":  saldo,
			})

			return nil
		})

		if err != nil {
			fmt.Println("Failed while acquiring db connection from pool")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		_, err = q.CreateTransacao(ctx, db.CreateTransacaoParams{
			Valor:     pgtype.Int8{Int64: req.Valor, Valid: true},
			Tipo:      pgtype.Text{String: req.Tipo, Valid: true},
			Descricao: pgtype.Text{String: req.Descricao, Valid: true},
			ClienteID: pgtype.Int8{Int64: id, Valid: true},
		})

		if err != nil {
			fmt.Println("Failed creating transacao")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		return
	}
}

func GetExtrato(ctx context.Context, q *db.Queries) func(*gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		transacoes, err := q.GetExtrato(ctx, pgtype.Int8{Int64: id, Valid: true})
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var resultTransacoes []interface{}
		var clienteSaldo int64
		var clienteLimite int64
		for _, transacao := range transacoes {
			resultTransacoes = append(resultTransacoes, gin.H{
				"valor":        transacao.Valor,
				"tipo":         transacao.Tipo,
				"descricao":    strings.Trim(transacao.Descricao.String, " "),
				"realizada_em": transacao.CreatedAt,
			})
			clienteSaldo = transacao.Saldo.Int64
			clienteLimite = transacao.Limite.Int64
		}

		response := gin.H{
			"saldo": gin.H{
				"data_extrato": time.Now().Format(time.RFC3339Nano),
				"total":        clienteSaldo,
				"limite":       clienteLimite,
			},
			"ultimas_transacoes": resultTransacoes,
		}

		c.JSON(http.StatusOK, response)
		return
	}
}
