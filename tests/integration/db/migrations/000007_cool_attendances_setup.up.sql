SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "cool_attendances" (
    id UUID PRIMARY KEY,
    cool_meeting_id UUID NOT NULL,
    community_id VARCHAR(15) NOT NULL,
    is_present BOOLEAN DEFAULT FALSE,
    remarks TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Index for looking up attendances by meeting
CREATE INDEX idx_cool_attendances_meeting_id ON cool_attendances(cool_meeting_id);

-- Index for looking up attendances by community
CREATE INDEX idx_cool_attendances_community_id ON cool_attendances(community_id);

-- Index for soft delete queries (used in your existing query)
CREATE INDEX idx_cool_attendances_deleted_at ON cool_attendances(deleted_at);

-- Composite index for common query pattern: finding attendances for a specific meeting in a specific community
CREATE INDEX idx_cool_attendances_meeting_community ON cool_attendances(cool_meeting_id, community_id);

-- Assuming cool_meetings table exists (which I can see it does in your migrations)
-- ALTER TABLE cool_attendances ADD CONSTRAINT fk_cool_attendances_meeting_id 
--     FOREIGN KEY (cool_meeting_id) REFERENCES cool_meetings(id);