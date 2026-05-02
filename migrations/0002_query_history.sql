-- +goose Up
CREATE TABLE IF NOT EXISTS "query_history" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  "query" TEXT NOT NULL,
  "executed_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_query_history_executed_at ON "query_history" ("executed_at");

-- +goose Down
DROP TABLE IF EXISTS "query_history";
