package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/db"
)

type PostTransacoesBody struct {
	Valor     int64
	Tipo      string
	Descricao string
}

type CreditDebitRow struct {
	Saldo  pgtype.Int8
	Limite pgtype.Int8
}

// This is ugly, but works
func creditDebit(ctx context.Context, q *db.Queries, tipo string, id int64, valor pgtype.Int8) (CreditDebitRow, error) {
	ret := CreditDebitRow{}
	var retErr error
	if tipo == "d" {
		row, err := q.Debit(ctx, db.DebitParams{ID: id, Saldo: valor})
		ret.Limite = row.Limite
		ret.Saldo = row.Saldo
		retErr = err
	} else if tipo == "c" {
		row, err := q.Credit(ctx, db.CreditParams{ID: id, Saldo: valor})
		ret.Limite = row.Limite
		ret.Saldo = row.Saldo
		retErr = err
	}
	return ret, retErr
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
			valor := pgtype.Int8{Int64: req.Valor, Valid: true}

			row, err := creditDebit(ctx, q, req.Tipo, id, valor)
			if err != nil && err.Error() == "no rows in result set" {
				c.AbortWithStatus(http.StatusUnprocessableEntity)
				return nil
			} else if err != nil {
				fmt.Println("Error updating cliente saldo")
				c.AbortWithError(http.StatusInternalServerError, err)
				return nil
			}

			c.JSON(http.StatusOK, gin.H{
				"limite": row.Limite.Int64,
				"saldo":  row.Saldo.Int64,
			})

			_, err = q.CreateTransacao(ctx, db.CreateTransacaoParams{
				Valor:     pgtype.Int8{Int64: req.Valor, Valid: true},
				Tipo:      pgtype.Text{String: req.Tipo, Valid: true},
				Descricao: pgtype.Text{String: req.Descricao, Valid: true},
				ClienteID: pgtype.Int8{Int64: id, Valid: true},
			})

			if err != nil {
				fmt.Println("Failed creating transacao")
				c.AbortWithStatus(http.StatusInternalServerError)
				return nil
			}

			return nil
		})

		if err != nil {
			fmt.Println("Failed while acquiring db connection from pool")
			c.AbortWithError(http.StatusInternalServerError, err)
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
		var clienteSaldo, clienteLimite int64
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

		// Deals with empty extrato
		if len(resultTransacoes) == 0 {
			cliente, err := q.GetCliente(ctx, id)
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			clienteSaldo = cliente.Saldo.Int64
			clienteLimite = cliente.Limite.Int64
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
