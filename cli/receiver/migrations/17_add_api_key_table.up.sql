CREATE TABLE api_key (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    hash TEXT NOT NULL
);