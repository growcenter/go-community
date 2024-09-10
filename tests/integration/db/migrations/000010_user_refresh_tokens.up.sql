SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "user_refresh_tokens" (
    "id" SERIAL PRIMARY KEY,
    "token_hash" TEXT NOT NULL,      -- Hashed refresh token
    "user_id" INT NOT NULL,          -- Foreign key to users table
    "expires_at" TIMESTAMP NOT NULL, -- Expiration date
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
