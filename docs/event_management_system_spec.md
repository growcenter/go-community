

# Church Event Management System - Technical Specification

**Version:** 3.0
**Date:** February 2026 (Revised - Removed event_occurrences table)
**Status:** ✅ Approved - Ready for Implementation

---

## 0. Design Decisions Summary

| Decision         | Choice                                         | Notes                                                         |
| ---------------- | ---------------------------------------------- | ------------------------------------------------------------- |
| Campus Filtering | Configurable per event                         | Can enable/disable campus restrictions                        |
| Migration        | Migration script                               | Transform existing data to new schema                         |
| Church Volunteer | Existing `volunteer` user_type                 | No changes needed                                             |
| Event Volunteer  | New `event_volunteer` role                     | To be created for per-event permissions                       |
| QR Generation    | Frontend generates                             | Backend provides data (community_id, registration_code, etc.) |
| Custom Forms     | Hierarchical (Event → Session → Child Session) | Each level can override parent                                |
| Notifications    | Tiered approach                                | Per pricing research below                                    |

---

## 1. Executive Summary

This document outlines the technical specification for a **modular, extensible church event management system**. The system is designed to handle diverse event types including Sunday Services, Christmas Celebrations, Conferences with multiple tracks, Volunteer Meetings, and Announcement-only events.

### Key Objectives

- **Modularity**: Flexible event structure supporting various use cases
- **Scalability**: Designed for free-tier constraints while being upgrade-ready
- **User Experience**: Streamlined QR-based registration and check-in flows
- **Data-Driven**: Comprehensive attendance tracking and reporting

### Technology Stack

| Component    | Technology            | Tier       |
| ------------ | --------------------- | ---------- |
| Backend      | Go (Echo Framework)   | Existing   |
| Database     | Supabase PostgreSQL   | Free Tier  |
| File Storage | Uploadthing           | Free Tier  |
| Hosting      | Google Cloud Platform | Free Tier  |
| Auth         | Existing RBAC System  | Integrated |

---

## 2. User Roles & Permissions

### 2.1 Role Hierarchy

```mermaid
graph TD
    A[Admin] --> B[Organizer]
    B --> C[Event Volunteer]
    B --> D[Church Volunteer]
    E[Regular Member] --> F[Guest]
```

### 2.2 Role Definitions

| Role                 | Description             | Key Capabilities                                                                                            |
| -------------------- | ----------------------- | ----------------------------------------------------------------------------------------------------------- |
| **Admin**            | System administrators   | Full CRUD on all events, manage users/roles, access all reports, configure system settings                  |
| **Organizer**        | Event creators/managers | Create/edit/delete own events, manage registrations, view reports for their events, assign event volunteers |
| **Event Volunteer**  | Per-event helpers       | Scan QR codes for check-in, view attendee list for assigned events, cannot modify event details             |
| **Church Volunteer** | Internal church staff   | Access volunteer-only private events, similar to Regular Member for public events                           |
| **Regular Member**   | Authenticated users     | Browse public events, register, view personal attendance history, manage profile                            |
| **Guest**            | Unauthenticated users   | View public announcements, limited registration (requires identifier)                                       |

### 2.3 Permission Matrix

| Action              | Admin | Organizer | Event Vol | Church Vol | Member | Guest |
| ------------------- | :---: | :-------: | :-------: | :--------: | :----: | :---: |
| Create Event        |  ✅   |    ✅     |    ❌     |     ❌     |   ❌   |  ❌   |
| Edit Own Event      |  ✅   |    ✅     |    ❌     |     ❌     |   ❌   |  ❌   |
| Delete Event        |  ✅   |   ✅\*    |    ❌     |     ❌     |   ❌   |  ❌   |
| View All Events     |  ✅   |    ✅     |    ❌     |     ❌     |   ❌   |  ❌   |
| View Public Events  |  ✅   |    ✅     |    ✅     |     ✅     |   ✅   |  ✅   |
| View Private Events |  ✅   |   ✅\*    |   ✅\*    |    ✅\*    |   ❌   |  ❌   |
| Scan QR (Check-in)  |  ✅   |    ✅     |    ✅     |     ❌     |   ❌   |  ❌   |
| Register for Events |  ✅   |    ✅     |    ✅     |     ✅     |   ✅   | ✅\*  |
| View Reports        |  ✅   |   ✅\*    |    ❌     |     ❌     |   ❌   |  ❌   |
| Export Data         |  ✅   |   ✅\*    |    ❌     |     ❌     |   ❌   |  ❌   |
| Manage Users        |  ✅   |    ❌     |    ❌     |     ❌     |   ❌   |  ❌   |

_\* = Limited to assigned/own resources_

---

## 3. Event Architecture

### 3.1 Entity Hierarchy

```mermaid
graph TD
    subgraph "Event Container"
        E[Event]
    end

    subgraph "Scheduling Layer"
        S[Session - Services/Classes/Tracks]
        S2[Child Session - Optional Hierarchy]
        RP[Recurrence Pattern - JSONB]
    end

    subgraph "Participation Layer"
        R[Registration Record]
        A[Attendance Record]
    end

    E --> S
    E --> RP
    S --> S2
    S --> R
    S2 --> R
    R --> A
```

> [!IMPORTANT]
> **On-Demand Occurrence Calculation**: For recurring events, occurrence dates are calculated on-demand based on the `recurrence_pattern` JSONB field. No separate `event_occurrences` table is used. This approach:
> 
> - Avoids pre-generating potentially hundreds of database rows
> - Improves performance for long-running recurring events
> - Makes recurrence pattern modifications instant (no regeneration needed)
> - Reduces database storage requirements

> [!NOTE]
> **Unified Session Model**: Sessions now handle all sub-event types (services, classes, tracks, breakouts). Use `session_type` to differentiate and `parent_session_id` for hierarchical events like conferences.

### 3.2 Event Types

| Type           | Description                        | Example                     | Registration | Check-in |
| -------------- | ---------------------------------- | --------------------------- | ------------ | -------- |
| `REGISTRATION` | Standard registerable event        | Christmas Event, Conference | Required     | Optional |
| `ATTENDANCE`   | Recurring attendance tracking      | Sunday Service              | Optional     | Required |
| `ANNOUNCEMENT` | Information-only                   | Event Announcement          | None         | None     |
| `VOLUNTEER`    | Internal volunteer events          | Monthly Volunteer Meeting   | Required     | Optional |
| `HYBRID`       | Combines registration + attendance | Training Series             | Required     | Required |

### 3.3 Event Visibility

| Level             | Who Can See                | Use Case                  |
| ----------------- | -------------------------- | ------------------------- |
| `PUBLIC`          | Everyone including guests  | Sunday Service, Christmas |
| `MEMBERS_ONLY`    | Authenticated members      | Member-exclusive events   |
| `VOLUNTEER_ONLY`  | Church Volunteers + Staff  | Internal volunteer events |
| `PRIVATE`         | Invited users only         | Leadership meetings       |
| `CAMPUS_SPECIFIC` | Users from specific campus | Campus-specific events    |

### 3.4 Registration Configuration

Sessions can be configured with different registration rules to handle various scenarios:

#### Registration Mode

| Mode                  | Description                                           | Use Case                                    |
| --------------------- | ----------------------------------------------------- | ------------------------------------------- |
| `self_only`           | User can only register themselves                     | Volunteer meetings, limited capacity events |
| `self_and_registered` | User can register themselves + other registered users | Family events where all must have accounts  |
| `self_and_others`     | User can register themselves + guests (default)       | Christmas service, public events            |

#### Additional Registration Options

| Option                            | Type    | Description                                                   |
| --------------------------------- | ------- | ------------------------------------------------------------- |
| `max_registrations_per_user`      | INTEGER | Max people one user can register (including self). Default: 1 |
| `one_session_per_event`           | BOOLEAN | If true, user can only register for ONE session in this event |
| `additional_registrant_form_mode` | ENUM    | Form requirements for additional registrants                  |

#### Form Modes for Additional Registrants

| Mode              | Primary Registrant               | Additional Registrants                         |
| ----------------- | -------------------------------- | ---------------------------------------------- |
| `same_as_primary` | Name + email/phone + custom form | Name + email/phone + same custom form          |
| `name_only`       | Name + email/phone + custom form | **Name only** (simplest - no contact required) |
| `custom`          | Name + email/phone + custom form | Uses `additional_registrant_form_id`           |

#### Example Configurations

```json
// Volunteer meeting: self-only, one session per event
{
  "registration_mode": "self_only",
  "max_registrations_per_user": 1,
  "one_session_per_event": true
}

// Christmas service: family registration allowed, friends just need name
{
  "registration_mode": "self_and_others",
  "max_registrations_per_user": 5,
  "additional_registrant_form_mode": "name_only"
}

// Conference: registered users only, separate form for guests
{
  "registration_mode": "self_and_registered",
  "max_registrations_per_user": 3,
  "additional_registrant_form_mode": "custom",
  "additional_registrant_form_id": 123
}
```

### 3.5 Identifier Configuration

Each session can configure which identifier fields are **visible** and **required** for both primary and additional registrants.

#### Available Identifier Fields

| Field      | Description             | Example            |
| ---------- | ----------------------- | ------------------ |
| `name`     | Full name of registrant | "John Doe"         |
| `email`    | Email address           | "john@example.com" |
| `phone`    | Phone number            | "+62812345678"     |
| `legal_id` | Legal ID (KTP/Passport) | "3171234567890001" |

#### Configuration Structure

Each field has two properties:

- **`visible`**: Whether the field is shown on the form
- **`required`**: Whether the field must be filled (only applies if visible)

```json
{
  "identifier_config": {
    "primary": {
      "name": { "visible": true, "required": true },
      "email": { "visible": true, "required": true },
      "phone": { "visible": true, "required": false },
      "legal_id": { "visible": false, "required": false }
    },
    "additional": {
      "name": { "visible": true, "required": true },
      "email": { "visible": false, "required": false },
      "phone": { "visible": false, "required": false },
      "legal_id": { "visible": false, "required": false }
    }
  }
}
```

#### Common Patterns

| Scenario                     | Primary Config                       | Additional Config  |
| ---------------------------- | ------------------------------------ | ------------------ |
| **Family registration**      | name ✓, email ✓, phone optional      | name ✓ only        |
| **Conference (all contact)** | name ✓, email ✓, phone ✓             | name ✓, email ✓    |
| **High-security event**      | name ✓, email ✓, phone ✓, legal_id ✓ | name ✓, legal_id ✓ |
| **Simple attendance**        | name ✓                               | name ✓             |

### 3.6 Geolocation Validation

Events and sessions can enforce geolocation validation to ensure attendees are physically present at the venue. This is configurable based on registration method.

#### Geolocation Configuration

| Field                          | Type    | Description                               |
| ------------------------------ | ------- | ----------------------------------------- |
| `geolocation_required`         | BOOLEAN | Whether geolocation validation is enabled |
| `geolocation_radius_meters`    | INTEGER | Allowed distance from venue (in meters)   |
| `geolocation_validation_rules` | JSONB   | Rules for when to validate location       |

#### Validation Rules by Registration Method

| Registration Method                 | Default Behavior                  | Configurable |
| ----------------------------------- | --------------------------------- | ------------ |
| **Personal QR** (scanned by staff)  | Offline - requires location check | ✅ Yes       |
| **Event/Session QR** (self-service) | Online - no location check        | ✅ Yes       |
| **Manual entry**                    | Offline - requires location check | ✅ Yes       |
| **Web registration**                | Online - no location check        | ✅ Yes       |

#### Configuration Structure

```json
{
  "geolocation_config": {
    "enabled": true,
    "venue_latitude": -6.2,
    "venue_longitude": 106.816666,
    "radius_meters": 100,
    "validation_rules": {
      "personal_qr": {
        "require_location": true,
        "allow_override": false
      },
      "event_qr": {
        "require_location": false,
        "allow_override": true
      },
      "session_qr": {
        "require_location": false,
        "allow_override": true
      },
      "manual": {
        "require_location": true,
        "allow_override": true
      },
      "web": {
        "require_location": false,
        "allow_override": false
      }
    },
    "error_action": "reject" // or "warn"
  }
}
```

#### Common Scenarios

| Scenario                | Configuration                                                   |
| ----------------------- | --------------------------------------------------------------- |
| **Strict on-site only** | All methods require location, radius: 50m, error_action: reject |
| **Hybrid event**        | Personal QR requires location, Event QR doesn't                 |
| **Flexible check-in**   | Location required but allow_override: true for staff            |
| **Fully online**        | geolocation_config.enabled: false                               |

#### Validation Flow

```
1. User attempts registration/check-in
2. System checks registration_method
3. If geolocation_config.validation_rules[method].require_location == true:
   a. Request user's current location
   b. Calculate distance from venue
   c. If distance > radius_meters:
      - If error_action == "reject": Block registration
      - If error_action == "warn": Allow but log warning
      - If allow_override == true: Staff can manually approve
4. Proceed with registration/check-in
```

---

## 4. Sequence Diagrams

### 4.1 Event Creation Flow

This diagram shows the complete flow of creating an event with all its components.

```mermaid
sequenceDiagram
    participant Client
    participant API as Event API
    participant UC as Event UseCase
    participant Validator
    participant Repo as Event Repository
    participant DB as Database

    Client->>API: POST /api/v1/events
    Note over Client,API: CreateEventRequest with:<br/>- Basic Info<br/>- Location<br/>- Schedule<br/>- Recurrence Pattern<br/>- Registration Config<br/>- Forms

    API->>Validator: Validate Request
    Validator->>Validator: Check required fields
    Validator->>Validator: Validate recurrence pattern
    Validator->>Validator: Validate location based on type
    Validator-->>API: Validation Result

    alt Validation Failed
        API-->>Client: 400 Bad Request
    end

    API->>UC: Create(ctx, request)

    UC->>UC: Generate Event Code
    UC->>UC: Set Default Timezone
    UC->>UC: Process Recurrence Pattern

    Note over UC: If recurrence pattern exists,<br/>validate and prepare for storage

    UC->>Repo: CreateEvent(event)
    Repo->>DB: BEGIN TRANSACTION

    Repo->>DB: INSERT INTO events
    DB-->>Repo: Event ID

    alt Has Recurrence Pattern
        Repo->>DB: Store recurrence_pattern JSONB
        Note over Repo,DB: Stores: frequency, interval,<br/>count, endDate, weekDays,<br/>monthlyPattern, excludeDates,<br/>additionalDates<br/><br/>Occurrences calculated on-demand<br/>when querying events
    end

    alt Has Registration Form
        Repo->>DB: Link registration_form_id
    end

    Repo->>DB: COMMIT TRANSACTION
    Repo-->>UC: Created Event

    UC-->>API: Event Response
    API-->>Client: 201 Created (Event Response)
```

### 4.2 Event Session Creation Flow

This diagram shows how sessions are created within an event.

```mermaid
sequenceDiagram
    participant Client
    participant API as Session API
    participant UC as Session UseCase
    participant Validator
    participant Repo as Session Repository
    participant DB as Database

    Client->>API: POST /api/v1/events/:code/sessions
    Note over Client,API: CreateSessionRequest with:<br/>- Session Info<br/>- Location<br/>- Schedule<br/>- Capacity<br/>- Check-in/Check-out Config<br/>- Registration Rules<br/>- Forms

    API->>Validator: Validate Request
    Validator->>Validator: Check session type
    Validator->>Validator: Validate capacity
    Validator->>Validator: Validate check-in config
    Validator->>Validator: Validate registration mode

    alt Check-in Validation
        Note over Validator: If check_in_required = true,<br/>must have time windows
        Note over Validator: If check_out_required = true,<br/>check_in must be enabled
    end

    Validator-->>API: Validation Result

    alt Validation Failed
        API-->>Client: 400 Bad Request
    end

    API->>UC: CreateSession(ctx, request)

    UC->>UC: Generate Session Code
    UC->>UC: Validate Parent Event Exists

    alt Parent Event Not Found
        UC-->>API: 404 Not Found
        API-->>Client: 404 Not Found
    end

    UC->>UC: Process Check-in/Check-out Config
    Note over UC: Set defaults:<br/>- check_in_enabled: true<br/>- check_in_window_before: 30<br/>- check_in_window_after: 15<br/>- check_in_late_threshold: 10

    UC->>UC: Process Form Configuration
    Note over UC: Handle:<br/>- registration_form_id<br/>- additional_registrant_form_mode<br/>- form_field_overrides

    UC->>Repo: CreateSession(session)
    Repo->>DB: BEGIN TRANSACTION

    Repo->>DB: INSERT INTO event_sessions
    Note over Repo,DB: Stores all fields including:<br/>- Check-in config (12 fields)<br/>- Check-out config (6 fields)<br/>- Registration rules<br/>- Form references

    alt Has Geolocation Config
        Repo->>DB: Store geolocation_config JSONB
    end

    alt Has Identifier Config
        Repo->>DB: Store identifier_config JSONB
    end

    alt Has Form Field Overrides
        Repo->>DB: Store form_field_overrides JSONB
    end

    Repo->>DB: COMMIT TRANSACTION
    Repo-->>UC: Created Session

    UC-->>API: Session Response
    API-->>Client: 201 Created (Session Response)
```

### 4.3 Recurring Event with Sessions Flow

This diagram shows the complete flow when creating a recurring event with multiple sessions.

```mermaid
sequenceDiagram
    participant Client
    participant EventAPI as Event API
    participant EventUC as Event UseCase
    participant SessionAPI as Session API
    participant SessionUC as Session UseCase
    participant DB as Database

    Client->>EventAPI: POST /api/v1/events
    Note over Client,EventAPI: Recurring event request<br/>with recurrence pattern

    EventAPI->>EventUC: Create(ctx, request)
    EventUC->>DB: INSERT INTO events
    DB-->>EventUC: Event Created

    Note over EventUC,DB: Recurrence pattern stored as JSONB.<br/>When querying events for a date range,<br/>occurrences are calculated on-the-fly<br/>based on this pattern.

    EventUC-->>EventAPI: Event Response
    EventAPI-->>Client: 201 Created

    Note over Client: Now create sessions<br/>for this event

    Client->>SessionAPI: POST /api/v1/events/:code/sessions
    Note over Client,SessionAPI: Kids Session (9:00-10:00)

    SessionAPI->>SessionUC: CreateSession(ctx, request)
    SessionUC->>DB: INSERT INTO event_sessions
    Note over SessionUC,DB: Session applies to<br/>all calculated occurrences

    DB-->>SessionUC: Kids Session Created
    SessionUC-->>SessionAPI: Session Response
    SessionAPI-->>Client: 201 Created

    Client->>SessionAPI: POST /api/v1/events/:code/sessions
    Note over Client,SessionAPI: Youth Session (10:00-11:00)

    SessionAPI->>SessionUC: CreateSession(ctx, request)
    SessionUC->>DB: INSERT INTO event_sessions
    DB-->>SessionUC: Youth Session Created
    SessionUC-->>SessionAPI: Session Response
    SessionAPI-->>Client: 201 Created

    Note over Client,DB: Result:<br/>1 Event<br/>2 Sessions<br/>Occurrences calculated on-demand
```

### 4.4 Registration with Check-in Flow

This diagram shows how check-in configuration affects the registration and attendance flow.

```mermaid
sequenceDiagram
    participant User
    participant RegAPI as Registration API
    participant AttAPI as Attendance API
    participant UC as UseCase
    participant DB as Database
    participant Validator as Check-in Validator

    Note over User,DB: Session Config:<br/>check_in_enabled: true<br/>check_in_required: true<br/>check_in_window_before: 30<br/>check_in_window_after: 15<br/>check_in_late_threshold: 10

    User->>RegAPI: POST /api/v1/registrations
    Note over User,RegAPI: Register for session<br/>Event starts at 09:00

    RegAPI->>UC: CreateRegistration(request)
    UC->>DB: INSERT INTO registrations
    DB-->>UC: Registration Created
    UC-->>RegAPI: Registration Response
    RegAPI-->>User: 201 Created + QR Code

    Note over User: Event day arrives<br/>User arrives at 08:45<br/>(15 min early)

    User->>AttAPI: POST /api/v1/attendance/check-in
    Note over User,AttAPI: Scan QR code<br/>Current time: 08:45

    AttAPI->>Validator: ValidateCheckInTime(session, currentTime)

    Validator->>Validator: Calculate check-in window
    Note over Validator: Window start: 08:30<br/>(30 min before)<br/>Window end: 09:15<br/>(15 min after)

    Validator->>Validator: Check if within window
    Note over Validator: 08:45 is within window ✓

    Validator->>Validator: Calculate lateness
    Note over Validator: Event starts: 09:00<br/>Check-in: 08:45<br/>Difference: -15 min<br/>Status: on_time

    Validator-->>AttAPI: Validation Passed

    AttAPI->>UC: RecordCheckIn(registration)
    UC->>DB: INSERT INTO attendance_records
    Note over UC,DB: check_in_at: 08:45<br/>status: on_time<br/>late_minutes: 0

    DB-->>UC: Attendance Recorded
    UC-->>AttAPI: Check-in Success
    AttAPI-->>User: 200 OK (Checked In)

    Note over User: Alternative: User arrives late<br/>at 09:12 (12 min late)

    User->>AttAPI: POST /api/v1/attendance/check-in
    AttAPI->>Validator: ValidateCheckInTime(session, 09:12)

    Validator->>Validator: Check if within window
    Note over Validator: 09:12 is within window ✓<br/>(before 09:15)

    Validator->>Validator: Calculate lateness
    Note over Validator: 09:12 - 09:00 = 12 min<br/>Late threshold: 10 min<br/>Status: late

    Validator-->>AttAPI: Validation Passed (Late)

    AttAPI->>UC: RecordCheckIn(registration)
    UC->>DB: INSERT INTO attendance_records
    Note over UC,DB: check_in_at: 09:12<br/>status: late<br/>late_minutes: 12

    DB-->>UC: Attendance Recorded
    UC-->>AttAPI: Check-in Success (Late)
    AttAPI-->>User: 200 OK (Checked In - Late)

    Note over User: Alternative: User arrives too late<br/>at 09:20 (20 min late)

    User->>AttAPI: POST /api/v1/attendance/check-in
    AttAPI->>Validator: ValidateCheckInTime(session, 09:20)

    Validator->>Validator: Check if within window
    Note over Validator: 09:20 > 09:15<br/>Outside window!

    Validator->>Validator: Check allow_late flag
    Note over Validator: check_in_allow_late: true

    alt Allow Late = true
        Validator-->>AttAPI: Validation Passed (Very Late)
        AttAPI->>UC: RecordCheckIn(registration)
        UC->>DB: INSERT INTO attendance_records
        Note over UC,DB: status: very_late<br/>late_minutes: 20
        UC-->>AttAPI: Check-in Success
        AttAPI-->>User: 200 OK (Very Late)
    else Allow Late = false
        Validator-->>AttAPI: Validation Failed
        AttAPI-->>User: 400 Check-in Window Closed
    end
```

### 4.5 Bulk Event + Sessions Creation Flow

This diagram shows how to create an event with multiple sessions in a single API call.

```mermaid
sequenceDiagram
    participant Client
    participant API as Event API
    participant Validator
    participant UC as Event UseCase
    participant SessionUC as Session UseCase
    participant Repo as Event Repository
    participant SessionRepo as Session Repository
    participant DB as Database
    participant OccGen as Occurrence Generator

    Client->>API: POST /api/v1/events/bulk
    Note over Client,API: CreateEventWithSessionsRequest:<br/>- Event data<br/>- Sessions: [Kids, Youth, Adult]<br/>- Recurrence pattern<br/>- Forms

    API->>Validator: Validate Request
    Validator->>Validator: Validate event data

    loop For each session
        Validator->>Validator: Validate session data
        Validator->>Validator: Check session times don't overlap
        Validator->>Validator: Validate check-in/check-out config
    end

    Validator-->>API: Validation Result

    alt Validation Failed
        API-->>Client: 400 Bad Request
    end

    API->>UC: CreateEventWithSessions(ctx, request)

    UC->>DB: BEGIN TRANSACTION
    Note over UC,DB: All-or-nothing:<br/>If any step fails,<br/>rollback everything

    UC->>UC: Generate Event Code
    UC->>UC: Process Recurrence Pattern

    UC->>Repo: CreateEvent(event)
    Repo->>DB: INSERT INTO events
    DB-->>Repo: Event ID & Code
    Repo-->>UC: Event Created

    Note over UC,DB: Recurrence pattern stored.<br/>No pre-generation of occurrences.

    Note over UC: Now create all sessions<br/>in the same transaction

    loop For each session in request
        UC->>SessionUC: CreateSession(eventCode, sessionData)

        SessionUC->>SessionUC: Generate Session Code
        SessionUC->>SessionUC: Validate against event schedule
        SessionUC->>SessionUC: Process check-in/check-out config

        SessionUC->>SessionRepo: CreateSession(session)
        SessionRepo->>DB: INSERT INTO event_sessions
        Note over SessionRepo,DB: Session 1: Kids (9:00-10:00)<br/>Session 2: Youth (10:00-11:00)<br/>Session 3: Adult (11:00-12:00)

        DB-->>SessionRepo: Session Created
        SessionRepo-->>SessionUC: Session Created
        SessionUC-->>UC: Session Created
    end

    alt Any Session Creation Failed
        UC->>DB: ROLLBACK TRANSACTION
        UC-->>API: 500 Internal Server Error
        API-->>Client: 500 Error (Nothing Created)
    else All Sessions Created Successfully
        UC->>DB: COMMIT TRANSACTION
        UC-->>API: Event + Occurrences + Sessions
        API-->>Client: 201 Created
        Note over Client,API: Response includes:<br/>- Event details<br/>- 12 Occurrences<br/>- 3 Sessions<br/>- 36 Session-Occurrence combinations
    end
```

### 4.6 Bulk Form + Questions Creation Flow

This diagram shows how to create a form with multiple questions in a single API call.

```mermaid
sequenceDiagram
    participant Client
    participant API as Form API
    participant Validator
    participant UC as Form UseCase
    participant Repo as Form Repository
    participant DB as Database

    Client->>API: POST /api/v1/forms/bulk
    Note over Client,API: CreateFormWithFieldsRequest:<br/>- Form metadata<br/>- Fields: [<br/>  {dietary_preference},<br/>  {tshirt_size},<br/>  {special_needs},<br/>  {transportation},<br/>  ...<br/>]

    API->>Validator: Validate Request
    Validator->>Validator: Validate form metadata
    Validator->>Validator: Check form_type is valid

    loop For each field
        Validator->>Validator: Validate field_key is unique
        Validator->>Validator: Validate field_type
        Validator->>Validator: Validate display_order
        Validator->>Validator: Check conditional logic
    end

    alt Has duplicate field_key
        Validator-->>API: Validation Error
        API-->>Client: 400 Duplicate field_key
    end

    Validator-->>API: Validation Result

    API->>UC: CreateFormWithFields(ctx, request)

    UC->>DB: BEGIN TRANSACTION
    Note over UC,DB: All-or-nothing:<br/>If any field fails,<br/>rollback entire form

    UC->>UC: Generate Form Code
    UC->>UC: Validate field dependencies
    Note over UC: Check conditional logic:<br/>If field B depends on field A,<br/>ensure field A exists

    UC->>Repo: CreateForm(form)
    Repo->>DB: INSERT INTO forms
    Note over Repo,DB: code: FORM-CHRISTMAS-2026<br/>title: Christmas Registration<br/>form_type: event_registration<br/>status: draft

    DB-->>Repo: Form ID
    Repo-->>UC: Form Created (ID: 123)

    Note over UC: Now create all fields<br/>in the same transaction

    loop For each field in request (10 fields)
        UC->>UC: Process field configuration
        Note over UC: Field 1: dietary_preference<br/>Field 2: tshirt_size<br/>Field 3: special_needs<br/>...

        UC->>Repo: CreateFormField(formID, field)
        Repo->>DB: INSERT INTO form_fields
        Note over Repo,DB: form_id: 123<br/>field_key: dietary_preference<br/>field_type: multiselect<br/>label: Dietary Restrictions<br/>display_order: 1<br/>is_required: false<br/>options: [vegetarian, vegan, ...]

        DB-->>Repo: Field ID
        Repo-->>UC: Field Created

        alt Has Conditional Logic
            UC->>Repo: CreateFieldCondition(fieldID, condition)
            Repo->>DB: INSERT INTO form_field_conditions
            Note over Repo,DB: field_id: 456<br/>depends_on_field_id: 123<br/>condition_type: equals<br/>condition_value: "yes"
            DB-->>Repo: Condition Created
        end

        alt Has Validation Rules
            UC->>Repo: CreateFieldValidation(fieldID, rules)
            Repo->>DB: INSERT INTO form_field_validations
            Note over Repo,DB: field_id: 456<br/>validation_type: regex<br/>validation_value: "^[0-9]{10}$"<br/>error_message: "Invalid phone"
            DB-->>Repo: Validation Created
        end
    end

    alt Any Field Creation Failed
        UC->>DB: ROLLBACK TRANSACTION
        UC-->>API: 500 Internal Server Error
        API-->>Client: 500 Error (Nothing Created)
    else All Fields Created Successfully
        UC->>DB: COMMIT TRANSACTION

        UC->>UC: Build Form Response
        Note over UC: Aggregate:<br/>- Form metadata<br/>- All fields with options<br/>- All conditions<br/>- All validations

        UC-->>API: Form + Fields
        API-->>Client: 201 Created
        Note over Client,API: Response includes:<br/>- Form details<br/>- 10 Fields<br/>- Field options<br/>- Conditional logic<br/>- Validation rules
    end
```

### 4.7 Event + Sessions + Forms Creation Flow (Complete)

This diagram shows the most complex scenario: creating everything at once.

```mermaid
sequenceDiagram
    participant Client
    participant EventAPI as Event API
    participant FormAPI as Form API
    participant EventUC as Event UseCase
    participant FormUC as Form UseCase
    participant SessionUC as Session UseCase
    participant DB as Database
    participant OccGen as Occurrence Generator

    Note over Client: Admin wants to create:<br/>- 1 Event (recurring)<br/>- 2 Forms (Primary + Kids)<br/>- 3 Sessions (Kids, Youth, Adult)

    Client->>FormAPI: POST /api/v1/forms/bulk
    Note over Client,FormAPI: Create Primary Registrant Form<br/>with 10 questions

    FormAPI->>FormUC: CreateFormWithFields(request)
    FormUC->>DB: BEGIN TRANSACTION
    FormUC->>DB: INSERT INTO forms
    DB-->>FormUC: Form ID: 100

    loop 10 questions
        FormUC->>DB: INSERT INTO form_fields
    end

    FormUC->>DB: COMMIT TRANSACTION
    FormUC-->>FormAPI: Form 100 Created
    FormAPI-->>Client: 201 Created (Form 100)

    Client->>FormAPI: POST /api/v1/forms/bulk
    Note over Client,FormAPI: Create Kids Form<br/>with 5 questions

    FormAPI->>FormUC: CreateFormWithFields(request)
    FormUC->>DB: BEGIN TRANSACTION
    FormUC->>DB: INSERT INTO forms
    DB-->>FormUC: Form ID: 101

    loop 5 questions
        FormUC->>DB: INSERT INTO form_fields
    end

    FormUC->>DB: COMMIT TRANSACTION
    FormUC-->>FormAPI: Form 101 Created
    FormAPI-->>Client: 201 Created (Form 101)

    Note over Client: Now create event with sessions<br/>and link to forms

    Client->>EventAPI: POST /api/v1/events/bulk
    Note over Client,EventAPI: CreateEventWithSessionsRequest:<br/>- Event (recurring, 12 weeks)<br/>- Sessions: [<br/>  Kids (form_id: 101),<br/>  Youth (form_id: 100),<br/>  Adult (form_id: 100)<br/>]

    EventAPI->>EventUC: CreateEventWithSessions(request)
    EventUC->>DB: BEGIN TRANSACTION

    EventUC->>DB: INSERT INTO events
    DB-->>EventUC: Event Created

    EventUC->>OccGen: GenerateOccurrences(event)
    loop 12 occurrences
        OccGen->>DB: INSERT INTO event_occurrences
    end
    OccGen-->>EventUC: 12 Occurrences Created

    loop For each session (3 sessions)
        EventUC->>SessionUC: CreateSession(sessionData)

        SessionUC->>SessionUC: Validate form_id exists
        Note over SessionUC: Kids session → Form 101<br/>Youth session → Form 100<br/>Adult session → Form 100

        SessionUC->>DB: INSERT INTO event_sessions
        Note over SessionUC,DB: Links to form via<br/>registration_form_id

        DB-->>SessionUC: Session Created
        SessionUC-->>EventUC: Session Created
    end

    EventUC->>DB: COMMIT TRANSACTION
    EventUC-->>EventAPI: Event + Occurrences + Sessions
    EventAPI-->>Client: 201 Created

    Note over Client,DB: Final Result:<br/>- 2 Forms (100, 101)<br/>- 15 Form Fields (10 + 5)<br/>- 1 Event<br/>- 12 Occurrences<br/>- 3 Sessions<br/>- 36 Session-Occurrence combinations

    Note over Client: Users can now:<br/>1. Register for sessions<br/>2. Fill out appropriate forms<br/>3. Check-in on event day
```

---

## 5. Database Schema

For detailed database schema including all tables, indexes, constraints, triggers, and migration notes, see:

📄 **[Database Schema Documentation](./database_schema.md)**

### Quick Reference

**Core Event Tables:**

- `events` - Main event definitions with recurrence patterns
- `event_sessions` - Sessions, services, tracks, breakouts (hierarchical)

**Form System (Production-Ready):**

- `forms` - Form definitions with ENUM types and status
- `form_questions` - Questions with context-aware filtering (`mandatory_for`, `apply_for`)
- `form_answers` - Answers with flexible identifiers (supports guests + authenticated users)
- `form_associations` - Polymorphic many-to-many relationships

**Registration & Attendance:**

- `registrations` - User registrations with QR codes
- `attendance_records` - Check-in/check-out tracking with geolocation

**Supporting Tables:**

- `event_series` - Grouping related events
- `event_volunteers` - Volunteer assignments

### Key Features

- **Type Safety**: ENUM types for `form_status`, `question_type`, `entity_type`, `identifier_type`
- **Data Integrity**: Foreign key constraints with CASCADE/SET NULL
- **Performance**: Comprehensive indexes including GIN indexes for array fields
- **Soft Deletes**: `deleted_at` columns with partial indexes
- **Auto-Update Triggers**: Automatic `updated_at` timestamp management
- **Helper Functions**: `clone_form_from_template()`, `get_form_questions_for_context()`

See the [full schema document](./database_schema.md) for:

- Complete table definitions with all columns
- Foreign key constraints and indexes
- Auto-update triggers
- Helper functions with examples
- Form schema migration strategy
- Testing checklist

---

### 5.1 QR Code Types

| QR Type         | Contains                     | Use Case                                   |
| --------------- | ---------------------------- | ------------------------------------------ |
| **Personal QR** | `USER:{community_id}`        | User's permanent QR for quick reg/check-in |
| **Event QR**    | `EVENT:{event_code}`         | Posted at venue, opens event registration  |
| **Session QR**  | `SESSION:{session_code}`     | For direct session/class/track registration|
| **Ticket QR**   | `TICKET:{registration_code}` | Confirmation ticket for verification       |

### 5.2 Registration Flows

#### Flow 1: Personal QR Scan (Fastest)

```mermaid
sequenceDiagram
    participant U as User
    participant S as Staff (Scanner)
    participant API as Backend
    participant DB as Database

    Note over U,S: User presents Personal QR
    S->>API: Scan Personal QR (USER:CID123)
    API->>DB: Get user info + available events
    DB-->>API: User data + active events nearby
    API-->>S: Show registration options
    S->>API: Select event/session + confirm
    API->>DB: Create registration (status: verified)
    API->>DB: Create attendance (check_in_at: now)
    DB-->>API: Success
    API-->>S: ✅ Registered + Checked In
    API-->>U: (Optional) Send confirmation
````

#### Flow 2: Event/Session QR Scan (Self-Service)

```mermaid
sequenceDiagram
    participant U as User
    participant W as Web App
    participant API as Backend
    participant DB as Database

    Note over U,W: User scans Event/Session QR
    U->>W: Scan EVENT:XMS2024
    W->>API: Get event details
    API->>DB: Query event + sessions
    DB-->>API: Event data
    API-->>W: Event info + sessions
    W-->>U: Show event page

    alt User is Logged In
        U->>W: Select session + Register
        W->>API: Create registration (authenticated)
        API->>DB: Check capacity + Create registration
        DB-->>API: Registration created
        API-->>W: Ticket QR code
        W-->>U: ✅ Show ticket + confirmation
    else User is Guest
        U->>W: Fill guest info + Register
        W->>API: Create registration (guest)
        API->>DB: Create registration
        API-->>W: Ticket QR code
        W-->>U: ✅ Show ticket + send to email
    end
```

#### Flow 3: Ticket Verification (Day-of Check-in)

```mermaid
sequenceDiagram
    participant U as User
    participant S as Staff (Scanner)
    participant API as Backend
    participant DB as Database

    Note over U,S: User shows Ticket QR
    S->>API: Scan TICKET:REG123ABC
    API->>DB: Get registration + check status

    alt Status = confirmed
        API->>DB: Create attendance record (checked_in)
        DB-->>API: Success
        API-->>S: ✅ Valid - [Name] - [Session]
    else Status = pending
        API-->>S: ⚠️ Pending - Needs confirmation
    else Status = cancelled
        API-->>S: ❌ Cancelled registration
    else Already checked in
        API-->>S: ℹ️ Already checked in at [time]
    end
```

### 5.3 Attendance Status Logic

```
Event Start Time: 09:00

Check-in at 08:45 → STATUS: on_time (15 min early)
Check-in at 09:10 → STATUS: late (10 min late)
Check-in at 10:00 → STATUS: very_late (60 min late)
No check-in       → STATUS: absent
Permit submitted  → STATUS: permit (with reason)
```

**Configurable thresholds per event:**

```json
{
  "lateThresholdMinutes": 15,
  "veryLateThresholdMinutes": 30,
  "allowLateCheckIn": true,
  "requireExcuseForAbsent": true
}
```

---

## 6. API Design

### 6.1 Event Endpoints

```
# Events
POST   /api/v1/events                    # Create event
GET    /api/v1/events                    # List events (with filters)
GET    /api/v1/events/:code              # Get event details
PUT    /api/v1/events/:code              # Update event
DELETE /api/v1/events/:code              # Delete event
POST   /api/v1/events/:code/duplicate    # Duplicate event
POST   /api/v1/events/:code/publish      # Publish event

# Sessions (handles services, classes, tracks, breakouts)
POST   /api/v1/events/:code/sessions     # Create session
GET    /api/v1/events/:code/sessions     # List sessions
PUT    /api/v1/sessions/:code            # Update session
DELETE /api/v1/sessions/:code            # Delete session

# Child Sessions (for hierarchical events like conferences)
POST   /api/v1/sessions/:code/children   # Create child session (track, breakout)
GET    /api/v1/sessions/:code/children   # List child sessions

# Occurrences (for recurring events)
GET    /api/v1/events/:code/occurrences  # List occurrences
PUT    /api/v1/occurrences/:code         # Modify occurrence
POST   /api/v1/occurrences/:code/cancel  # Cancel occurrence
```

### 6.2 Registration Endpoints

```
# Registration
POST   /api/v1/registrations                      # Create registration
GET    /api/v1/registrations/:code                # Get registration
PUT    /api/v1/registrations/:code                # Update registration
DELETE /api/v1/registrations/:code                # Cancel registration
POST   /api/v1/registrations/:code/confirm        # Confirm registration

# QR Flows
POST   /api/v1/scan/personal                      # Scan personal QR
POST   /api/v1/scan/event                         # Scan event QR
POST   /api/v1/scan/ticket                        # Scan ticket QR

# My Registrations (User self-service)
GET    /api/v1/me/registrations                   # My registrations
GET    /api/v1/me/attendance                      # My attendance history
```

### 6.3 Attendance Endpoints

```
# Check-in/out
POST   /api/v1/attendance/check-in                # Manual check-in
POST   /api/v1/attendance/check-out               # Manual check-out
GET    /api/v1/events/:code/attendance            # Event attendance list
PUT    /api/v1/attendance/:id/status              # Update status (permit, excuse)

# Bulk Operations
POST   /api/v1/events/:code/attendance/bulk-check-in   # Bulk check-in
POST   /api/v1/events/:code/attendance/mark-absent     # Mark remaining as absent
```

### 6.4 Reports Endpoints

```
# Event Reports
GET    /api/v1/reports/events/:code/summary                # Event summary
GET    /api/v1/reports/events/:code/registrations          # Registration report
GET    /api/v1/reports/events/:code/attendance             # Attendance report
GET    /api/v1/reports/events/:code/export?format=csv|xlsx # Export data

# User Reports
GET    /api/v1/reports/users/:communityId/attendance       # User attendance history

# Dashboard
GET    /api/v1/reports/dashboard                           # Overview stats
GET    /api/v1/reports/trends                              # Attendance trends
```

---

## 7. Notification System

### 7.1 Recommended Providers (Paid - Cheapest Options)

#### Email Providers

| Provider       | Free Tier             | Paid Pricing           | Recommendation      |
| -------------- | --------------------- | ---------------------- | ------------------- |
| **Amazon SES** | 3,000/month (on EC2)  | $0.10 per 1,000 emails | ⭐ Best for scale   |
| **Resend**     | 100/day (3,000/month) | $20/mo for 50K emails  | Good for starting   |
| **Brevo**      | 300/day (9,000/month) | $15/mo for 20K emails  | Best free tier      |
| **Mailtrap**   | 4,000/month           | $15/mo for 10K emails  | Good deliverability |

#### SMS Providers

| Provider      | Per SMS (Indonesia) | Notes                      |
| ------------- | ------------------- | -------------------------- |
| **Bandwidth** | $0.004/msg          | Very competitive           |
| **Telnyx**    | $0.004/msg          | Transparent pricing        |
| **Twilio**    | $0.0079/msg         | Most popular, higher price |
| **Plivo**     | $0.002/msg (India)  | Cheapest for Asia          |

> [!TIP]
> **Recommended Starting Setup:**
> 
> 1. **Email**: Start with Brevo free tier (300/day = 9,000/month)
> 2. **SMS**: Skip initially, add Plivo when needed (~$20/month for 10K SMS)
> 3. **WhatsApp**: Use WhatsApp Business API via Twilio ($0.005/msg) when budget allows

### 7.2 Notification Types

| Type                      | Trigger                    | Channels | Priority |
| ------------------------- | -------------------------- | -------- | -------- |
| Registration Confirmation | On successful registration | Email    | High     |
| Event Reminder            | 24h + 1h before event      | Email    | Medium   |
| Waitlist Promotion        | When spot opens            | Email    | High     |
| Event Update              | When event details change  | In-App   | Low      |
| Cancellation Notice       | When event cancelled       | Email    | High     |

### 7.3 Implementation Strategy

```go
// Notification service interface
type NotificationService interface {
    SendEmail(ctx context.Context, to, subject, body string) error
    SendSMS(ctx context.Context, to, message string) error
    SendInApp(ctx context.Context, userID, message string) error
}

// Provider abstraction for easy switching
type EmailProvider interface {
    Send(email Email) error
}
```

Phased rollout:

1. **Phase 1**: In-app notifications only (no cost)
2. **Phase 2**: Add Brevo for email (free tier)
3. **Phase 3**: Add SMS when budget allows

---

## 8. Reporting & Analytics

### 8.1 Core Reports

#### Registration Report

| Metric                 | Description                       |
| ---------------------- | --------------------------------- |
| Total Registrations    | Count of all registrations        |
| Confirmed vs Pending   | Registration status breakdown     |
| By Session             | Distribution across sessions      |
| By Registration Method | Personal QR vs Event QR vs Manual |
| Registration Timeline  | Registrations over time           |

#### Attendance Report

| Metric          | Description                   |
| --------------- | ----------------------------- |
| Attendance Rate | Checked-in / Registered       |
| On-Time vs Late | Punctuality analysis          |
| No-Shows        | Registered but didn't attend  |
| Walk-Ins        | Attended without registration |
| By Session      | Attendance per session        |

#### Trend Analysis

| Report             | Scope                          |
| ------------------ | ------------------------------ |
| Weekly Attendance  | Sunday Service trends          |
| Monthly Volunteers | Volunteer meeting attendance   |
| Event Comparison   | Compare similar events         |
| User Engagement    | Individual attendance patterns |

### 8.2 Export Formats

| Format   | Use Case              | Implementation     |
| -------- | --------------------- | ------------------ |
| **CSV**  | Simple data export    | Native Go encoding |
| **XLSX** | Excel with formatting | `excelize` library |
| **PDF**  | Printable reports     | Future enhancement |

---

## 9. Free Tier Considerations

### 9.1 Supabase PostgreSQL Free Tier

| Limit                | Value          | Strategy                             |
| -------------------- | -------------- | ------------------------------------ |
| Database Size        | 500MB          | Optimize schemas, archive old data   |
| API Requests         | 500K/month     | Cache frequently accessed data       |
| Realtime Connections | 200 concurrent | Use polling for non-critical updates |
| Auth Users           | 50K MAU        | Sufficient for church context        |

### 9.2 GCP Free Tier

| Service         | Free Tier         | Usage              |
| --------------- | ----------------- | ------------------ |
| Cloud Run       | 2M requests/month | API hosting        |
| Cloud Storage   | 5GB               | QR code images     |
| Cloud Scheduler | 3 jobs            | Cron for reminders |

### 9.3 Uploadthing Free Tier

| Limit   | Value   | Usage                      |
| ------- | ------- | -------------------------- |
| Storage | 2GB     | Event images, user avatars |
| Uploads | 100/day | Sufficient for church use  |

---

## 10. Implementation Phases

### Phase 1: Core Event Management (Weeks 1-3)

- [ ] Database schema migrations
- [ ] Event CRUD API
- [ ] Session CRUD API
- [ ] Basic registration flow
- [ ] Personal QR code system

### Phase 2: QR & Check-in (Weeks 4-5)

- [ ] QR code generation service
- [ ] Scan API endpoints
- [ ] Check-in/check-out flow
- [ ] Web scanner interface

### Phase 3: Recurring Events (Week 6)

- [ ] Occurrence generation
- [ ] Attendance tracking for recurring
- [ ] Template system

### Phase 4: Custom Forms (Week 7)

- [ ] Form builder API
- [ ] Dynamic form rendering
- [ ] Form answer storage

### Phase 5: Reporting (Week 8)

- [ ] Core report APIs
- [ ] CSV/XLSX export
- [ ] User attendance history

### Phase 6: Notifications & Polish (Weeks 9-10)

- [ ] Email notifications
- [ ] Reminder system
- [ ] UI polish
- [ ] Performance optimization

---

## 11. File Structure (Go Project)

```
internal/
├── constants/
│   ├── event_constants.go       # Event types, statuses
│   └── registration_constants.go
├── models/
│   ├── event_model.go           # Event entity + DTOs
│   ├── session_model.go         # Session entity + DTOs (unified: services, classes, tracks)
│   ├── occurrence_model.go      # Occurrence entity + DTOs
│   ├── registration_model.go    # Registration entity + DTOs
│   ├── attendance_model.go      # Attendance entity + DTOs
│   ├── registration_form_model.go # Custom forms
│   └── qr_code_model.go         # QR code entities
├── repositories/
│   ├── pgsql/
│   │   ├── event_repository.go
│   │   ├── session_repository.go  # Unified session repository
│   │   ├── registration_repository.go
│   │   ├── attendance_repository.go
│   │   └── report_repository.go
│   └── interfaces.go
├── usecases/
│   ├── event/
│   │   ├── event_usecase.go
│   │   └── session_usecase.go  # Unified session usecase
│   ├── registration/
│   │   ├── registration_usecase.go
│   │   └── qr_scan_usecase.go
│   ├── attendance/
│   │   └── attendance_usecase.go
│   └── report/
│       └── report_usecase.go
├── deliveries/
│   └── http/
│       ├── event_handler.go
│       ├── session_handler.go
│       ├── registration_handler.go
│       ├── scan_handler.go
│       ├── attendance_handler.go
│       └── report_handler.go
├── services/
│   ├── qr_generator_service.go
│   ├── notification_service.go
│   └── export_service.go
└── pkg/
    ├── qrcode/
    │   └── qrcode.go            # QR generation utilities
    └── export/
        ├── csv_exporter.go
        └── xlsx_exporter.go

migrations/
├── 000001_create_events.up.sql
├── 000002_create_sessions.up.sql  # Unified sessions table
├── 000003_create_occurrences.up.sql
├── 000004_create_form_enums.up.sql
├── 000005_create_forms.up.sql
├── 000006_create_form_questions.up.sql
├── 000007_create_form_answers.up.sql
├── 000008_create_form_associations.up.sql
├── 000009_create_form_triggers.up.sql
├── 000010_create_form_helper_functions.up.sql
├── 000011_create_registrations.up.sql
├── 000012_create_attendance.up.sql
└── 000013_create_user_qr_codes.up.sql
```

---

## 12. Questions for Clarification

Before proceeding to implementation, please confirm:

1. **Campus System**: I see you have existing campus infrastructure. Should events be campus-aware (e.g., events visible/specific to certain campuses)?
2. **Existing Event System**: The current codebase has `events`, `event_instances`, and `event_registration_records`. Should we:
   
   - A) Completely replace these tables with the new schema?
   - B) Migrate data from old to new schema?
   - C) Keep both and deprecate old gradually?
3. **User Authentication**: The spec assumes integration with existing RBAC. Should the `Church Volunteer` role be:
   
   - A) A new role in the existing roles table?
   - B) A new user_type?
   - C) Both (role for permissions, user_type for classification)?
4. **Notification Priority**: Given the free tier limits, which notification should be prioritized?
   
   - Registration confirmation
   - Event reminders
   - Both (with batching)
5. **QR Code Storage**: Should QR code images be:
   
   - A) Generated on-demand and not stored
   - B) Pre-generated and stored in Uploadthing
   - C) Stored as base64 in database

---

## Appendix A: Example Event Configurations

### Sunday Service (Recurring Attendance)

```json
{
  "title": "Sunday Service",
  "category": "attendance",
  "is_recurring": true,
  "recurrence_pattern": {
    "type": "weekly",
    "days": ["sunday"],
    "interval": 1
  },
  "sessions": [
    { "title": "Service 1", "start_at": "07:00", "end_at": "09:00" },
    { "title": "Service 2", "start_at": "09:30", "end_at": "11:30" },
    { "title": "Service 3", "start_at": "17:00", "end_at": "19:00" }
  ],
  "check_in_required": true,
  "check_out_required": false,
  "allow_guest_registration": false
}
```

### Christmas Event (Multi-Session Registration)

```json
{
  "title": "Christmas Celebration 2026",
  "category": "registration",
  "sessions": [
    {
      "title": "General Service",
      "session_type": "general",
      "capacity": 500,
      "start_at": "2026-12-25T09:00:00"
    },
    {
      "title": "Kids Service",
      "session_type": "kids",
      "capacity": 100,
      "start_at": "2026-12-25T09:00:00",
      "age_config": { "min_age": 3, "max_age": 12 }
    }
  ],
  "notification_channels": ["email"],
  "check_in_required": true
}
```

### Conference (Hierarchical Sessions with Tracks)

```json
{
  "title": "Leadership Conference 2026",
  "category": "registration",
  "sessions": [
    {
      "title": "Day 1",
      "session_type": "general",
      "children": [
        {
          "title": "Track A: Leadership",
          "session_type": "track",
          "capacity": 50
        },
        {
          "title": "Track B: Communication",
          "session_type": "track",
          "capacity": 50
        },
        {
          "title": "Track C: Worship",
          "session_type": "workshop",
          "capacity": 30
        }
      ]
    },
    {
      "title": "Day 2",
      "session_type": "general",
      "children": [
        {
          "title": "Track A: Advanced Leadership",
          "session_type": "track",
          "capacity": 50
        },
        {
          "title": "Track B: Team Building",
          "session_type": "track",
          "capacity": 50
        }
      ]
    }
  ],
  "registration_form": {
    "fields": [
      {
        "type": "select",
        "label": "Dietary Preference",
        "options": ["None", "Vegetarian", "Halal"]
      }
    ]
  }
}
```

### Announcement Event

```json
{
  "title": "New Year's Eve Service Announcement",
  "category": "announcement",
  "description": "Join us for our special New Year's Eve service!",
  "virtual_link": "https://youtube.com/live/xyz",
  "call_to_action": {
    "text": "Watch Live",
    "url": "https://youtube.com/live/xyz"
  },
  "notification_channels": ["email"]
}
```

---

---

## 10. API Request/Response Examples

### 10.1 Create One-Time Event

**Request:**

```http
POST /api/v1/events
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Christmas Celebration 2026",
  "slug": "christmas-2026",
  "description": "Join us for our annual Christmas celebration!",
  "category": "registration",
  "status": "active",
  
  "images": {
    "bannerLink": "https://cdn.example.com/christmas-banner.jpg",
    "imageLinks": [
      "https://cdn.example.com/christmas-1.jpg",
      "https://cdn.example.com/christmas-2.jpg"
    ]
  },
  
  "location": {
    "locationType": "hybrid",
    "physicalAddress": "Church Main Building, 123 Main St, Jakarta",
    "physicalPlaceName": "Main Sanctuary",
    "virtualLink": "https://youtube.com/live/christmas2026",
    "virtualPlatform": "youtube",
    "locationVisibility": "all"
  },
  
  "schedule": {
    "startAt": "2026-12-25T09:00:00+07:00",
    "endAt": "2026-12-25T12:00:00+07:00",
    "timezone": "Asia/Jakarta"
  },
  
  "access": {
    "accessLevel": "public"
  },
  
  "organizer": {
    "organizerCommunityIds": ["comm_abc123", "comm_def456"]
  }
}
```

**Response:**

```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "type": "event",
  "code": "EVT-202612-A3K9M",
  "id": 12345,
  "title": "Christmas Celebration 2026",
  "slug": "christmas-2026",
  "category": "registration",
  "status": "active",
  "isRecurring": false,
  "location": {
    "locationType": "hybrid",
    "physicalAddress": "Church Main Building, 123 Main St, Jakarta",
    "virtualLink": "https://youtube.com/live/christmas2026"
  },
  "schedule": {
    "startAt": "2026-12-25T09:00:00+07:00",
    "endAt": "2026-12-25T12:00:00+07:00",
    "timezone": "Asia/Jakarta"
  },
  "createdAt": "2026-02-10T13:00:00+07:00",
  "updatedAt": "2026-02-10T13:00:00+07:00"
}
```

### 10.2 Create Recurring Event (Weekly Sunday Service)

**Request:**

```http
POST /api/v1/events
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Sunday Service",
  "description": "Weekly Sunday worship service",
  "category": "attendance",
  "status": "active",
  
  "location": {
    "locationType": "hybrid",
    "physicalAddress": "Church Main Building",
    "virtualLink": "https://youtube.com/@church/live",
    "virtualPlatform": "youtube",
    "locationVisibility": "all"
  },
  
  "schedule": {
    "startAt": "2026-02-16T09:00:00+07:00",
    "endAt": "2026-02-16T11:00:00+07:00",
    "timezone": "Asia/Jakarta"
  },
  
  "recurrence": {
    "isRecurring": true,
    "recurrencePattern": {
      "frequency": "weekly",
      "interval": 1,
      "weekDays": ["sunday"],
      "count": 52,
      "excludeDates": ["2026-03-15", "2026-04-20"]
    }
  },
  
  "access": {
    "accessLevel": "public"
  },
  
  "organizer": {
    "organizerCommunityIds": ["comm_abc123"]
  }
}
```

**Response:**

```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "type": "event",
  "code": "EVT-202602-B7X2P",
  "id": 12346,
  "title": "Sunday Service",
  "category": "attendance",
  "status": "active",
  "isRecurring": true,
  "recurrencePattern": {
    "frequency": "weekly",
    "interval": 1,
    "weekDays": ["sunday"],
    "count": 52,
    "excludeDates": ["2026-03-15", "2026-04-20"]
  },
  "schedule": {
    "startAt": "2026-02-16T09:00:00+07:00",
    "endAt": "2026-02-16T11:00:00+07:00",
    "timezone": "Asia/Jakarta"
  },
  "nextOccurrences": [
    "2026-02-16T09:00:00+07:00",
    "2026-02-23T09:00:00+07:00",
    "2026-03-02T09:00:00+07:00",
    "2026-03-09T09:00:00+07:00",
    "2026-03-16T09:00:00+07:00"
  ],
  "createdAt": "2026-02-10T13:00:00+07:00"
}
```

### 10.3 Query Events with Date Range

**Request:**

```http
GET /api/v1/events?startDate=2026-02-16&endDate=2026-03-16&category=attendance
Authorization: Bearer <token>
```

**Response:**

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "events": [
    {
      "code": "EVT-202602-B7X2P",
      "title": "Sunday Service",
      "category": "attendance",
      "isRecurring": true,
      "occurrencesInRange": [
        {
          "date": "2026-02-16",
          "startAt": "2026-02-16T09:00:00+07:00",
          "endAt": "2026-02-16T11:00:00+07:00"
        },
        {
          "date": "2026-02-23",
          "startAt": "2026-02-23T09:00:00+07:00",
          "endAt": "2026-02-23T11:00:00+07:00"
        },
        {
          "date": "2026-03-02",
          "startAt": "2026-03-02T09:00:00+07:00",
          "endAt": "2026-03-02T11:00:00+07:00"
        },
        {
          "date": "2026-03-09",
          "startAt": "2026-03-09T09:00:00+07:00",
          "endAt": "2026-03-09T11:00:00+07:00"
        }
      ]
    }
  ],
  "pagination": {
    "total": 1,
    "page": 1,
    "perPage": 20
  }
}
```

> [!NOTE]
> The `occurrencesInRange` array is calculated on-demand based on the recurrence pattern and the requested date range. No database rows are pre-generated.

### 10.4 Create Event Session

**Request:**

```http
POST /api/v1/events/EVT-202602-B7X2P/sessions
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Kids Service",
  "description": "Sunday service for children ages 3-12",
  "sessionType": "kids",
  
  "schedule": {
    "startAt": "2026-02-16T09:00:00+07:00",
    "endAt": "2026-02-16T10:30:00+07:00"
  },
  
  "location": {
    "locationType": "offline",
    "physicalAddress": "Kids Building, Room 101",
    "roomName": "Rainbow Room"
  },
  
  "capacity": {
    "capacity": 50,
    "waitlistEnabled": true,
    "waitlistCapacity": 20
  },
  
  "registration": {
    "registrationMode": "self_and_others",
    "maxRegistrationsPerUser": 5,
    "additionalRegistrantFormMode": "name_only"
  },
  
  "checkIn": {
    "enabled": true,
    "required": true,
    "windowBefore": 30,
    "windowAfter": 15,
    "allowLate": true,
    "lateThreshold": 10
  },
  
  "ageConfig": {
    "minAge": 3,
    "maxAge": 12
  }
}
```

**Response:**

```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "type": "event_session",
  "code": "EVT-202602-B7X2P-K1D5",
  "id": 5001,
  "eventCode": "EVT-202602-B7X2P",
  "title": "Kids Service",
  "sessionType": "kids",
  "capacity": 50,
  "currentCount": 0,
  "checkInConfig": {
    "enabled": true,
    "required": true,
    "windowBefore": 30,
    "windowAfter": 15
  },
  "createdAt": "2026-02-10T13:00:00+07:00"
}
```

### 10.5 Register for Recurring Event Session

**Request:**

```http
POST /api/v1/registrations
Content-Type: application/json
Authorization: Bearer <token>

{
  "eventCode": "EVT-202602-B7X2P",
  "sessionCode": "EVT-202602-B7X2P-K1D5",
  "occurrenceDate": "2026-02-16",
  "registrants": [
    {
      "isPrimary": true,
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+628123456789"
    },
    {
      "isPrimary": false,
      "name": "Jane Doe"
    },
    {
      "isPrimary": false,
      "name": "Jimmy Doe"
    }
  ]
}
```

**Response:**

```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "registrationCode": "REG-202602-X9K2M",
  "eventCode": "EVT-202602-B7X2P",
  "sessionCode": "EVT-202602-B7X2P-K1D5",
  "occurrenceDate": "2026-02-16",
  "registrants": [
    {
      "name": "John Doe",
      "qrCode": "data:image/png;base64,iVBORw0KG..."
    },
    {
      "name": "Jane Doe",
      "qrCode": "data:image/png;base64,iVBORw0KG..."
    },
    {
      "name": "Jimmy Doe",
      "qrCode": "data:image/png;base64,iVBORw0KG..."
    }
  ],
  "createdAt": "2026-02-10T13:00:00+07:00"
}
```

---

_End of Technical Specification_