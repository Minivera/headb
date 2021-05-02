CREATE TABLE "collections" (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id BIGSERIAL NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES "users"(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX collections_user_id_name_unique_index ON "collections"(name, user_id);

ALTER TABLE "collections" ADD CONSTRAINT name_user_id_unique UNIQUE USING INDEX collections_user_id_name_unique_index;