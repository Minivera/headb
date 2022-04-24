CREATE TABLE "api_keys" (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    value VARCHAR(60) NOT NULL,
    user_id UUID DEFAULT uuid_generate_v4() NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES "users"(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX api_key_value_user_id_unique_index ON "api_keys"(value, user_id);

ALTER TABLE "api_keys" ADD CONSTRAINT value_user_id_unique UNIQUE USING INDEX api_key_value_user_id_unique_index;