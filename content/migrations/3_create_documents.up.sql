CREATE TABLE "documents" (
   id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
   content jsonb NOT NULL DEFAULT '{}'::jsonb,
   collection_id UUID NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   CONSTRAINT fk_collection FOREIGN KEY(collection_id) REFERENCES "collections"(id) ON DELETE CASCADE
);