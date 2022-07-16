CREATE TABLE "transactions" (
   id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
   script TEXT NOT NULL,
   database_id UUID NOT NULL,
   expires_at TIMESTAMPTZ NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   CONSTRAINT fk_database FOREIGN KEY(database_id) REFERENCES "databases"(id) ON DELETE CASCADE
);

CREATE TABLE "transactional_collections" (
   id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
   name VARCHAR(255) NOT NULL,
   database_id UUID NOT NULL,
   transaction_id UUID NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   CONSTRAINT fk_database FOREIGN KEY(database_id) REFERENCES "databases"(id) ON DELETE CASCADE,
   CONSTRAINT fk_transaction FOREIGN KEY(transaction_id) REFERENCES "transactions"(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX transactional_collections_database_id_name_unique_index ON "transactional_collections"(name, database_id);

ALTER TABLE "transactional_collections" ADD CONSTRAINT transaction_name_database_id_unique UNIQUE USING INDEX transactional_collections_database_id_name_unique_index;

CREATE TABLE "transactional_documents" (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    content jsonb NOT NULL DEFAULT '{}'::jsonb,
    collection_id UUID NOT NULL,
    transaction_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_transaction FOREIGN KEY(transaction_id) REFERENCES "transactions"(id) ON DELETE CASCADE
);