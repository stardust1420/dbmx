-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS "connections" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  "engine" VARCHAR NOT NULL,
  "host" VARCHAR NOT NULL,
  "port" VARCHAR NOT NULL,
  "username" VARCHAR NOT NULL,
  "password" VARCHAR NOT NULL,
  "database" VARCHAR NOT NULL,
  "name" VARCHAR NOT NULL,
  "env" VARCHAR DEFAULT NULL,
  "colour" VARCHAR DEFAULT NULL,
  "is_advanced" INTEGER NOT NULL DEFAULT 0,
  "ssl_mode" VARCHAR DEFAULT NULL,
  "client_key" BLOB DEFAULT NULL,
  "server_cert" BLOB DEFAULT NULL,
  "root_ca_cert" BLOB DEFAULT NULL,
  "over_ssh" INTEGER NOT NULL DEFAULT 0,
  "ssh_host" VARCHAR DEFAULT NULL,
  "ssh_port" VARCHAR DEFAULT NULL,
  "ssh_username" VARCHAR DEFAULT NULL,
  "ssh_password" VARCHAR DEFAULT NULL,
  "use_ssh_key" INTEGER NOT NULL DEFAULT 0,
  "ssh_key" BLOB DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS "tabs" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL,
  "editor" TEXT,
  "output" TEXT,
  "is_active" INTEGER NOT NULL DEFAULT 0,
  "active_db_id" VARCHAR,
  "active_db" VARCHAR,
  "active_db_colour" VARCHAR,
  "type" VARCHAR NOT NULL DEFAULT 'editor',
  "postgres_conn_id" BIGINT DEFAULT NULL,
  "db_name" VARCHAR DEFAULT NULL,
  "postgres_conn_name" VARCHAR NOT NULL DEFAULT '',
  "select" TEXT NOT NULL DEFAULT '',
  "limit" VARCHAR NOT NULL DEFAULT '',
  "offset" VARCHAR NOT NULL DEFAULT '',
  "where" TEXT NOT NULL DEFAULT '',
  "order_by" TEXT NOT NULL DEFAULT '',
  "group_by" TEXT NOT NULL DEFAULT '',
  "table_columns" TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_tabs_is_active ON "tabs" ("is_active");

-- +goose Down
DROP TABLE IF EXISTS "tabs";
DROP TABLE IF EXISTS "postgres";
