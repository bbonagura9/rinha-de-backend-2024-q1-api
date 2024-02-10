package routes

import (
  "errors"
  "fmt"
  "net/http"

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
      Valor: req.Valor,
      Tipo: req.Tipo,
      Descricao: req.Descricao,
    })

    c.JSON(http.StatusOK, gin.H{
      "limite": cliente.Limite,
      "saldo": cliente.Saldo,
    })
  }
}
