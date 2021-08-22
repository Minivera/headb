CREATE TYPE user_status AS ENUM ('pending', 'accepted', 'denied');

ALTER TABLE "users" ALTER COLUMN username DROP NOT NULL;
ALTER TABLE "users" ADD unique_id varchar(55);
ALTER TABLE "users" ADD status user_status;

CREATE UNIQUE INDEX user_unique_id_unique_index ON "users"(unique_id);

ALTER TABLE "users" ADD CONSTRAINT user_unique_id_unique UNIQUE USING INDEX user_unique_id_unique_index;