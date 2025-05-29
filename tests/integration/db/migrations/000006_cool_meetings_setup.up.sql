SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "cool_meetings" (
    id UUID PRIMARY KEY,
    cool_code VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    meeting_date DATE NOT NULL,
    meeting_start_at TIME NOT NULL,
    meeting_end_at TIME NOT NULL,
    new_joiners TEXT[],
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_cool_meetings_cool_code ON cool_meetings(cool_code);
