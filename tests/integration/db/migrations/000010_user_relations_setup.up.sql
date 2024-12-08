SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "user_relations" (
    "id" SERIAL PRIMARY KEY,
    "master_community_id" varchar(15) NOT NULL UNIQUE,
    "spouse_community_id" varchar(15) UNIQUE,
    "children_community_id" TEXT[],
    "relation_community_id" TEXT[],
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);