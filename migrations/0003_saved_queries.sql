-- +goose Up
CREATE TABLE IF NOT EXISTS "saved_queries" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  "title" TEXT NOT NULL,
  "query" TEXT NOT NULL,
  "saved_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_saved_queries_saved_at ON "saved_queries" ("saved_at");

-- +goose Down
DROP TABLE IF EXISTS "saved_queries";
