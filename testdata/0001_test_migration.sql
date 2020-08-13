-- This is a general comment for the entire migration file
-- package: foo
-- migrate: up
-- NOTE: this is a comment for the up migration
-- Comments should be included but ignored

CREATE TABLE IF NOT EXISTS users (
    "id" integer PRIMARY KEY,
    "username" varchar(128) NOT NULL UNIQUE,
    "email" varchar(128) NOT NULL UNIQUE,
    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS groups (
    "id" integer PRIMARY KEY,
    "name" varchar(255) NOT NULL UNIQUE,
    "active" boolean DEFAULT false,
);

CREATE TABLE IF NOT EXISTS membership (
    "id" integer PRIMARY KEY,
    "user_id" integer NOT NULL,
    "group_id" integer NOT NULL,
    FOREIGN KEY ("user_id") REFERENCES users("id"),
    FOREIGN KEY ("group_id") REFERENCES groups("id"),
    UNIQUE("user_id", "group_id")
);

-- migrate: down
-- NOTE: this is a comment for the down migration

DROP TABLE IF EXISTS users CASCADE;