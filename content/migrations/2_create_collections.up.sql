CREATE TABLE "collections" (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    database_id BIGSERIAL NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_collection FOREIGN KEY(database_id) REFERENCES "databases"(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX collections_database_id_name_unique_index ON "collections"(name, database_id);

ALTER TABLE "collections" ADD CONSTRAINT name_database_id_unique UNIQUE USING INDEX collections_database_id_name_unique_index;