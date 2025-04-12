SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "cool_new_joiners" (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    gender varchar(6) NOT NULL,
    marital_status varchar(50) NOT NULL,
    year_of_birth INT NOT NULL,
    phone_number varchar(15) NOT NULL,
    address TEXT NOT NULL,
    campus_code varchar(3) NOT NULL,
    location varchar(100) NOT NULL,
    community_of_interest varchar(100) NOT NULL,
    updated_by varchar(150),
    status varchar(30) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP -- Nullable (for soft delete)
);