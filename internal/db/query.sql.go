// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createTransacao = `-- name: CreateTransacao :one
INSERT INTO transacoes (
  valor, tipo, descricao, cliente_id
) VALUES (
  $1, $2, $3, $4
) RETURNING id, created_at, valor, tipo, descricao, cliente_id
`

type CreateTransacaoParams struct {
	Valor     pgtype.Int8
	Tipo      pgtype.Text
	Descricao pgtype.Text
	ClienteID pgtype.Int8
}

func (q *Queries) CreateTransacao(ctx context.Context, arg CreateTransacaoParams) (Transaco, error) {
	row := q.db.QueryRow(ctx, createTransacao,
		arg.Valor,
		arg.Tipo,
		arg.Descricao,
		arg.ClienteID,
	)
	var i Transaco
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.Valor,
		&i.Tipo,
		&i.Descricao,
		&i.ClienteID,
	)
	return i, err
}

const credit = `-- name: Credit :one
UPDATE clientes SET
  saldo = saldo + $2
WHERE id = $1
RETURNING saldo, limite
`

type CreditParams struct {
	ID    int64
	Saldo pgtype.Int8
}

type CreditRow struct {
	Saldo  pgtype.Int8
	Limite pgtype.Int8
}

func (q *Queries) Credit(ctx context.Context, arg CreditParams) (CreditRow, error) {
	row := q.db.QueryRow(ctx, credit, arg.ID, arg.Saldo)
	var i CreditRow
	err := row.Scan(&i.Saldo, &i.Limite)
	return i, err
}

const debit = `-- name: Debit :one
UPDATE clientes SET
  saldo = saldo - $2
WHERE id = CASE 
  WHEN saldo - $2 < -limite THEN -1
  ELSE $1 END
RETURNING saldo, limite
`

type DebitParams struct {
	ID    int64
	Saldo pgtype.Int8
}

type DebitRow struct {
	Saldo  pgtype.Int8
	Limite pgtype.Int8
}

func (q *Queries) Debit(ctx context.Context, arg DebitParams) (DebitRow, error) {
	row := q.db.QueryRow(ctx, debit, arg.ID, arg.Saldo)
	var i DebitRow
	err := row.Scan(&i.Saldo, &i.Limite)
	return i, err
}

const getCliente = `-- name: GetCliente :one
SELECT id, created_at, limite, saldo FROM clientes
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetCliente(ctx context.Context, id int64) (Cliente, error) {
	row := q.db.QueryRow(ctx, getCliente, id)
	var i Cliente
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.Limite,
		&i.Saldo,
	)
	return i, err
}

const getClienteLock = `-- name: GetClienteLock :one
SELECT id, created_at, limite, saldo FROM clientes
WHERE id = $1 LIMIT 1
FOR UPDATE
`

func (q *Queries) GetClienteLock(ctx context.Context, id int64) (Cliente, error) {
	row := q.db.QueryRow(ctx, getClienteLock, id)
	var i Cliente
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.Limite,
		&i.Saldo,
	)
	return i, err
}

const getExtrato = `-- name: GetExtrato :many
SELECT t.id, t.created_at, t.valor, t.tipo, t.descricao, t.cliente_id, c.saldo, c.limite FROM transacoes t
LEFT JOIN clientes c
ON c.id = t.cliente_id
WHERE t.cliente_id = $1
ORDER BY t.created_at DESC
LIMIT 10
`

type GetExtratoRow struct {
	ID        int64
	CreatedAt pgtype.Timestamptz
	Valor     pgtype.Int8
	Tipo      pgtype.Text
	Descricao pgtype.Text
	ClienteID pgtype.Int8
	Saldo     pgtype.Int8
	Limite    pgtype.Int8
}

func (q *Queries) GetExtrato(ctx context.Context, clienteID pgtype.Int8) ([]GetExtratoRow, error) {
	rows, err := q.db.Query(ctx, getExtrato, clienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetExtratoRow
	for rows.Next() {
		var i GetExtratoRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.Valor,
			&i.Tipo,
			&i.Descricao,
			&i.ClienteID,
			&i.Saldo,
			&i.Limite,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
