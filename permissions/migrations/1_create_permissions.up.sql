CREATE TYPE role AS ENUM ('admin', 'write', 'read');

CREATE TABLE "permissions" (
    id BIGSERIAL PRIMARY KEY,
    key_id BIGINT NOT NULL,
    database_id BIGINT,
    role role NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX permissions_key_database_id_unique_index ON "permissions"(key_id, database_id);

ALTER TABLE "permissions" ADD CONSTRAINT key_database_id_unique UNIQUE USING INDEX permissions_key_database_id_unique_index;
