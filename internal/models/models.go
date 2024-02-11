package models

import (
	"errors"

	"gorm.io/gorm"
)

type Transacao struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey"`
	Valor     uint64 `gorm:"name:valor"`
	Tipo      string `gorm:"name:tipo" gorm:"size:2"`
	Descricao string `gorm:"name:descricao" gorm:"size:16"`
	ClienteID uint   `gorm:"name:cliente_id"`
}

func (Transacao) TableName() string {
	return "transacoes"
}

type Cliente struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey"`
	Limite     uint64 `gorm:"name:limite"`
	Saldo      int64  `gorm:"name:saldo"`
	Transacoes []Transacao
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
		if err := db.First(&cliente, idx).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			cliente = Cliente{
				ID:     idx,
				Limite: limites[idx],
				Saldo:  0,
			}
			db.Create(&cliente)
		}
	}
}
