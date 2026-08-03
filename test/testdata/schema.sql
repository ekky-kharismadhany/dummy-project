CREATE TABLE IF NOT EXISTS pokemon_lookup_log (
    id serial PRIMARY KEY,
    name text NOT NULL,
    fetched_at timestamptz NOT NULL
);
