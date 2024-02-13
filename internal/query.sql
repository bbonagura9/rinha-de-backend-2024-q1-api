-- name: GetClienteLock :one
SELECT * FROM clientes
WHERE id = $1 LIMIT 1
FOR UPDATE;

-- name: GetCliente :one
SELECT * FROM clientes
WHERE id = $1 LIMIT 1;

-- name: CreateTransacao :one
INSERT INTO transacoes (
  valor, tipo, descricao, cliente_id
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetExtrato :many
SELECT t.*, c.saldo, c.limite FROM transacoes t
LEFT JOIN clientes c
ON c.id = t.cliente_id
WHERE t.cliente_id = $1
ORDER BY t.created_at DESC
LIMIT 10;

-- name: Debit :one
UPDATE clientes SET
  saldo = saldo - $2
WHERE id = CASE 
  WHEN saldo - $2 < -limite THEN -1
  ELSE $1 END
RETURNING saldo, limite;

-- name: Credit :one
UPDATE clientes SET
  saldo = saldo + $2
WHERE id = $1
RETURNING saldo, limite;

