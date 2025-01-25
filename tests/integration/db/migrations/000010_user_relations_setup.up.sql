SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "user_relations" (
    "id" SERIAL PRIMARY KEY,
    "community_id" varchar(15) NOT NULL REFERENCES users(community_id),
    "related_community_id" varchar(15) NOT NULL REFERENCES users(community_id),
    "relationship_type" VARCHAR(20) CHECK (relationship_type IN ('spouse', 'parent', 'child')),
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP,
    PRIMARY KEY (community_id, related_community_id)
);