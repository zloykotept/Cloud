CREATE TABLE IF NOT EXISTS files
(
	"id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	"name" TEXT NOT NULL,
	"size" REAL NOT NULL,
	"update" TIMESTAMP WITH TIME ZONE NOT NULL,
	"public" BOOLEAN NOT NULL,
	"favourite" BOOLEAN NOT NULL,
	"owner_id" BIGINT NOT NULL,
	"dir_id" UUID,
	FOREIGN KEY("owner_id") REFERENCES users("id") ON DELETE CASCADE,
	FOREIGN KEY("dir_id") REFERENCES dirs("id") ON DELETE CASCADE
);
CREATE EXTENSION IF NOT EXISTS "pgcrypto";