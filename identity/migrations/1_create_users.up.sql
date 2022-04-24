CREATE EXTENSION "uuid-ossp";

CREATE TABLE "users" (
   id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
   username VARCHAR(255) NOT NULL,
   token VARCHAR(255),
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX user_username_unique_index ON "users"(username);

ALTER TABLE "users" ADD CONSTRAINT user_username_unique UNIQUE USING INDEX user_username_unique_index;
