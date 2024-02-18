CREATE UNLOGGED TABLE clients (
    id SERIAL PRIMARY KEY,
    balance INTEGER NOT NULL,
    limit INTEGER NOT NULL
);

-- ALTER TABLE
--     clients DISABLE ROW LEVEL SECURITY;

---
CREATE UNLOGGED TABLE transactions (
    id SERIAL PRIMARY KEY,
    client_id INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    description VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

ALTER TABLE
    transactions DISABLE ROW LEVEL SECURITY;

CREATE INDEX IF NOT EXISTS transactions_client_id_idx ON transactions(client_id ASC);

---
DO $$ BEGIN
    INSERT INTO
        clientes (nome, limite)
    VALUES
        ('o barato sai caro', 1000 * 100),
        ('zan corp ltda', 800 * 100),
        ('les cruders', 10000 * 100),
        ('padaria joia de cocaia', 100000 * 100),
        ('kid mais', 5000 * 100);

END;

$$