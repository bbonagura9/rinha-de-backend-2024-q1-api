CREATE UNLOGGED TABLE clientes (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    created_at timestamp with time zone,
    limite bigint,
    saldo bigint
);

CREATE UNLOGGED TABLE transacoes (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    valor bigint,
    tipo char(1),
    descricao char(10),
    cliente_id bigint
);

