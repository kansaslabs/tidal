-- This table is used to track the state of migrations as different revisions are applied
-- migrate: up

CREATE TABLE IF NOT EXISTS migrations (
    "revision" integer NOT NULL,
    "name" varchar(128) NOT NULL,
    "active" boolean NOT NULL DEFAULT false,
    "applied" TIMESTAMP WITH TIME ZONE,
    "created" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("revision")
) WITHOUT OIDS;

COMMENT ON TABLE "migrations" IS 'Manages the state of database by enabling migrations and rollbacks';
COMMENT ON COLUMN "migrations"."revision" IS 'The revision id parsed from the filename of the migration';
COMMENT ON COLUMN "migrations"."name" IS 'The name of the migration parsed from the filename of the migration';
COMMENT ON COLUMN "migrations"."active" IS 'If the migration has been applied, set to false on rollbacks or if not applied';
COMMENT ON COLUMN "migrations"."applied" IS 'Timestamp when the migration was applied, null if rolledback or not applied';
COMMENT ON COLUMN "migrations"."created" IS 'Timestamp when the migration was created';

-- The down migration will take the database all the way back to a blank slate
-- migrate: down

DROP TABLE IF EXISTS migrations CASCADE;