CREATE EXTENSION "uuid-ossp";

CREATE TABLE "databases" (
   id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
   name VARCHAR(255) NOT NULL,
   user_id UUID NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX databases_user_id_name_unique_index ON "databases"(name, user_id);

ALTER TABLE "databases" ADD CONSTRAINT name_user_id_unique UNIQUE USING INDEX databases_user_id_name_unique_index;
