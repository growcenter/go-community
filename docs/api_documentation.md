# Event Management System - API Documentation

**Version:** 3.0  
**Last Updated:** February 2026  
**Related Documents:** [Event Management System Specification](./event_management_system_spec.md) | [Database Schema](./database_schema.md)

---

## Overview

This document contains the complete API documentation for the Event Management System, including all endpoints, request/response formats, and examples.

### API Conventions

- **Base URL**: `/api/v1`
- **Authentication**: Bearer token in `Authorization` header
- **Content-Type**: `application/json`
- **Date Format**: ISO 8601 with timezone (e.g., `2026-02-10T13:00:00+07:00`)
- **Pagination**: Query parameters `page` and `perPage`
- **Filtering**: Query parameters for field-specific filters

### Response Codes

| Code | Meaning |
|------|---------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource conflict (e.g., duplicate) |
| 500 | Internal Server Error |

---

## 1. Event Endpoints

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

## 2. Registration Endpoints

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

## 3. Attendance Endpoints

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

## 4. Reports Endpoints

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

## Detailed Request/Response Examples

### 1. Create One-Time Event

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

### 2. Create Recurring Event (Weekly Sunday Service)

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

### 3. Query Events with Date Range

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

### 4. Create Event Session

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

### 5. Register for Recurring Event Session

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
