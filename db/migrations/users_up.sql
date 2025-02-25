CREATE TABLE IF NOT EXISTS users (
    "id"	SERIAL,
    "name"	VARCHAR(30) NOT NULL UNIQUE,
    "password"	BYTEA NOT NULL,
    "permissions"	INTEGER NOT NULL,
    "space"	REAL NOT NULL,
    PRIMARY KEY("id")
);
CREATE INDEX IF NOT EXISTS idx_name ON users (name);