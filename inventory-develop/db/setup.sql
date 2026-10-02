-- for create purchase table

CREATE TABLE purchases (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    quantity JSONB NOT NULL,
    current_price NUMERIC NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
