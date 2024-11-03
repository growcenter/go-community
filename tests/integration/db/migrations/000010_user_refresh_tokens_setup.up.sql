SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "user_refresh_tokens" (
    "id" SERIAL PRIMARY KEY,
    "token_hash" TEXT NOT NULL,
    "user_id" INT NOT NULL,
    "expires_at" TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);