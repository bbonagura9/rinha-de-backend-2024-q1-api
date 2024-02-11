package routes

import (
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func PostTrancacoes(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var req models.TransacoesPostBody
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

		id := c.Param("id")
		var cliente models.Cliente
		db.Transaction(func(tx *gorm.DB) error {
			result := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&cliente, id)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.AbortWithStatus(http.StatusNotFound)
				return nil
			}

			if req.Tipo == "d" {
				cliente.Saldo = cliente.Saldo - int64(req.Valor)
				if cliente.Saldo < -int64(cliente.Limite) {
					c.AbortWithStatus(http.StatusUnprocessableEntity)
					return nil
				}
			} else if req.Tipo == "c" {
				cliente.Saldo = cliente.Saldo + int64(req.Valor)
			}

			tx.Save(&cliente)
			tx.Create(&models.Transacao{
				Valor:     req.Valor,
				Tipo:      req.Tipo,
				Descricao: req.Descricao,
				ClienteID: cliente.ID,
			})

			c.JSON(http.StatusOK, gin.H{
				"limite": cliente.Limite,
				"saldo":  cliente.Saldo,
			})

			return nil
		})
	}
}

func GetExtrato(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		var cliente models.Cliente

		result := db.First(&cliente, id)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var transacoes []models.Transacao
		cond := map[string]interface{}{"cliente_id": cliente.ID}
		if result := db.Limit(10).Order("updated_at desc").Find(&transacoes, cond); result.Error != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var resultTransacoes []interface{}
		for _, transacao := range transacoes {
			resultTransacoes = append(resultTransacoes, gin.H{
				"valor":        transacao.Valor,
				"tipo":         transacao.Tipo,
				"descricao":    transacao.Descricao,
				"realizada_em": transacao.UpdatedAt,
			})
		}

		response := gin.H{
			"saldo": gin.H{
				"total":        cliente.Saldo,
				"data_extrato": time.Now().Format(time.RFC3339Nano),
				"limite":       cliente.Limite,
			},
			"ultimas_transacoes": resultTransacoes,
		}

		c.JSON(http.StatusOK, response)
		return
	}
}
