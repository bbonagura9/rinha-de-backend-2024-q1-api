package routes

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/bbonagura9/rinha-de-backend-2024-q1-api/internal/models"
	"github.com/gin-gonic/gin"
)

func PostTrancacoes(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		var cliente models.Cliente
		result := db.First(&cliente, id)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var req models.TransacoesPostBody
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if req.Tipo == "d" {
			cliente.Saldo = cliente.Saldo - int64(req.Valor)
			if cliente.Saldo < -int64(cliente.Limite) {
				c.AbortWithStatus(http.StatusUnprocessableEntity)
				return
			}
		} else if req.Tipo == "c" {
			cliente.Saldo = cliente.Saldo + int64(req.Valor)
		}
		fmt.Println(cliente.Saldo)

		db.Save(&cliente)
		db.Create(&models.Transacao{
			Valor:     req.Valor,
			Tipo:      req.Tipo,
			Descricao: req.Descricao,
			ClienteID: cliente.ID,
		})

		c.JSON(http.StatusOK, gin.H{
			"limite": cliente.Limite,
			"saldo":  cliente.Saldo,
		})
	}
}

func GetExtrato(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		var cliente models.Cliente

		if result := db.First(&cliente, id); errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var transacoes []models.Transacao
		cond := map[string]interface{}{"cliente_id": cliente.ID}
		if result := db.Limit(10).Order("updated_at").Find(&transacoes, cond); result.Error != nil {
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

		result := gin.H{
			"saldo": gin.H{
				"total":        cliente.Saldo,
				"data_extrato": time.Now().Format(time.RFC3339Nano),
				"limite":       cliente.Limite,
			},
			"ultimas_transacoes": resultTransacoes,
		}

		c.JSON(http.StatusOK, result)
	}
}
