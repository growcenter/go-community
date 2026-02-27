SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "configs" (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    identifier varchar(255) NOT NULL,
    key varchar(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP -- Nullable (for soft delete)
);