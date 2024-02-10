package models

import (
  "gorm.io/gorm"
)

type Transacao struct {
  gorm.Model
  ID        uint   `gorm:"primaryKey"`
  Valor     uint64 `gorm:"name:valor"`
  Tipo      string `gorm:"name:tipo"`
  Descricao string `gorm:"name:descricao"`
}

func (Transacao) TableName() string {
  return "transacoes"
}

type Cliente struct {
  gorm.Model
  ID     uint   `gorm:"primaryKey"`
  Limite uint64 `gorm:"name:limite"`
  Saldo  int64  `gorm:"name:saldo"`
}

type TransacoesPostBody struct {
  Valor     uint64
  Tipo      string
  Descricao string
}

func InitClientes(db *gorm.DB) {
  limites := [6]uint64{
    0,
    100000,
    80000,
    1000000,
    10000000,
    500000,
  }
  for idx := uint(1); idx <= 5; idx++ {
    var cliente Cliente
    if result := db.First(&cliente, idx); result.RowsAffected == 0 {
      cliente = Cliente{
        ID: idx,
        Limite: limites[idx],
        Saldo: 0,
      }
      db.Create(&cliente)
    }
  }
}
