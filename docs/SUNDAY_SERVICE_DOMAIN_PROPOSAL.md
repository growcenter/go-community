__# Sunday Service Domain: Schema & Flow

This document outlines the proposed database schema and system flow for the new `Sunday Service` domain. This domain is designed to handle the scheduling and attendance of the serving team (volunteers and staff), while the existing `Event` system will continue to manage ticketing and registration for the general congregation.

---

## Proposed Database Schema

Four new tables will be created to manage serving team schedules, roles, assignments, and attendance.

```sql
-- Stores the actual dates of the Sunday services.
CREATE TABLE sunday_service_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_date DATE NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Defines the roles or positions that volunteers can be assigned to.
CREATE TABLE sunday_service_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    team VARCHAR(255), -- e.g., "Music", "Welcome", "Production"
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- This is the core table, linking a user to a specific role for a specific service date.
CREATE TABLE sunday_service_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    schedule_id UUID NOT NULL REFERENCES sunday_service_schedules(id),
    role_id UUID NOT NULL REFERENCES sunday_service_roles(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(user_id, schedule_id, role_id)
);

-- Tracks the attendance for each assignment on the day of the service.
CREATE TABLE sunday_service_attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL UNIQUE REFERENCES sunday_service_assignments(id),
    status VARCHAR(50) NOT NULL, -- e.g., "PRESENT", "ABSENT", "LATE"
    check_in_time TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## System Flow and API Endpoints

This section describes the user interaction with the system through API endpoints.

### 1. Setup (Admin Task)

-   **Create Roles:** An admin defines all the possible serving roles.
    -   **Endpoint:** `POST /api/v1/sunday-service/roles`
    -   **Action:** Creates records in the `sunday_service_roles` table (e.g., "Guitarist", "Vocalist", "Usher").
-   **Create Schedules:** An admin creates the service dates for the upcoming weeks or months.
    -   **Endpoint:** `POST /api/v1/sunday-service/schedules`
    -   **Action:** Creates records in the `sunday_service_schedules` table for each upcoming Sunday.

### 2. Scheduling (Admin Task)

-   **Assign Volunteers:** The admin assigns users to roles for specific dates.
    -   **Endpoint:** `POST /api/v1/sunday-service/assignments`
    -   **Action:** This is the main scheduling step. It creates a record in `sunday_service_assignments`, linking a `user_id`, `schedule_id`, and `role_id`.

### 3. Volunteer View

-   **Check Personal Schedule:** A volunteer logs in to see their upcoming assignments.
    -   **Endpoint:** `GET /api/v1/sunday-service/assignments?user_id={me}`
    -   **Action:** The system fetches all records from `sunday_service_assignments` for that user and joins them with `schedules` and `roles` to show the date and role for each assignment.

### 4. Day of Service (Attendance Tracking)

-   **Mark Attendance:** A service leader or the volunteer checks in.
    -   **Endpoint:** `POST /api/v1/sunday-service/attendance`
    -   **Body:** `{ "assignment_id": "...", "status": "PRESENT" }`
    -   **Action:** Creates a record in the `sunday_service_attendance` table linked to the assignment. The `status` can be "PRESENT", "LATE", or an admin can mark someone as "ABSENT".

### 5. Reporting (Admin Task)

-   **View Service Roster & Attendance:** An admin wants to see the full attendance sheet for a past service.
    -   **Endpoint:** `GET /api/v1/sunday-service/attendance?schedule_id={schedule_id}`
    -   **Action:** The system fetches all attendance data for the given schedule, joining across all four tables to produce a complete report showing each assigned person, their role, and their attendance status ("PRESENT", "ABSENT", "LATE").

---

## Hybrid Approach Summary

This new domain works in parallel with the existing `Event` system.

| System                | Manages                   | User                       | Purpose                               |
| --------------------- | ------------------------- | -------------------------- | ------------------------------------- |
| **`Event` System**    | General Attendees         | Public Congregation        | Ticketing, Registration, Seat Management |
| **`Sunday Service` Domain** | Serving Team / Volunteers | Ministry Leaders & Volunteers | Role Scheduling, Roster Management, Duty Attendance |
