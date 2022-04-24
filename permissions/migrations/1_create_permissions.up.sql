CREATE EXTENSION "uuid-ossp";

CREATE TYPE role AS ENUM ('admin', 'write', 'read');

CREATE TABLE "permissions" (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    key_id UUID NOT NULL,
    database_id UUID,
    role role NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX permissions_key_database_id_unique_index ON "permissions"(key_id, database_id);

ALTER TABLE "permissions" ADD CONSTRAINT key_database_id_unique UNIQUE USING INDEX permissions_key_database_id_unique_index;
