# Comprehensive Event System Refactoring & Issues Report

This document provides a detailed breakdown of all issues, architectural decisions, and improvements identified during the code review and refactoring of the Event Management System.

## 1. Critical Implementation Gaps (Missing Logic)

These are functional gaps in the `Create` usecase that must be addressed for the system to work according to the specificaton.

### 🔴 High Priority

1.  **Occurrence Generation for Recurring Events**
    - **Issue**: The `Create` function saves the event but **does not generate the individual occurrences** (instances) for recurring events.
    - **Impact**: Recurring events will appear in the database but won't show up on calendars or allow session management.
    - **Required Fix**: Call `occurrenceGenerator.GenerateOccurrences(ctx, event)` within the creation transaction.

2.  **Location Type Validation**
    - **Issue**: No business logic exists to enforce that "Online" events have a link and "Offline" events have an address.
    - **Required Fix**: Implement `validateLocationByType` to ensure data integrity based on `LocationType`.

3.  **Schedule Sanity Check**
    - **Issue**: `EndAt` is not strictly validated to be after `StartAt` in the business layer (only via struct tags which can be bypassed or fail with pointers).
    - **Required Fix**: Explicit check `if !EndAt.After(StartAt) { return error }`.

## 2. Event Model Refactoring (`internal/models/event_model.go`)

### 🐛 Critical Bugs Fixed

1.  **Broken Validation Syntax**
    - **Problem**: Used pipe `|` for chaining `required_if` rules (e.g., `required_if=A|required_if=B`). This is invalid in the validator library.
    - **Fix**: Changed to comma `,` (e.g., `required_if=A,required_if=B`).
    - **Code**:
      ```go
      // BEFORE
      validate:"required_if=LocationType offline|required_if=LocationType hybrid"
      // AFTER
      validate:"omitempty,required_if=LocationType offline,required_if=LocationType hybrid"
      ```

2.  **Import Cycle Error**
    - **Problem**: `models` package imports `internal/pkg/errorc`, which likely imports other packages that depend on `models`.
    - **Fix**: Remove the `errorc` dependency from `models`. Models should result in simple errors or booleans; rich error handling belongs in the Usecase layer.

3.  **Nil Pointer Panic Risks**
    - **Problem**: `ToCreateResponse` method was blindly dereferencing pointers (e.g., `*e.TermsAndConditions`).
    - **Fix**: Added nil checks before dereferencing.

### ✨ Enhancements

1.  **Security Limits**: Added `max` length tags to string fields to prevent DoS attacks (e.g., `max=255`, `max=2048` for URLs).
2.  **Type Safety**: Added `uuid4` validation for ID arrays and `timezone` validation for timezone strings.

## 3. Architectural Design Decisions

### 🏗️ Composition Over Duplication (Request Structs)

We discussed whether to separate request structs (e.g., `EventLocation`) for Create vs. Update operations.

- **Decision**: **Keep Shared Structs**
- **Reasoning**:
  1.  **Single Source of Truth**: Validation rules are defined once.
  2.  **Consistency**: API contracts remain consistent across lifecycle.
  3.  **Maintainability**: Changes to "Location" structure automatically propagate to both operations.

### 📍 Helper Methods Strategy

We discussed where helper logic (like `.IsPublic()`) should live.

- **Decision**: **Implement on BOTH Request DTOs and Domain Models**
- **Reasoning**:
  - **DTO Methods (`EventAccess.IsPublic`)**: Used during request binding/validation **before** the domain request is created.
  - **Model Methods (`Event.IsPublic`)**: Used for business logic, DB queries, and access control **after** persistence.
  - **Benefit**: Decouples API layer validation from core Domain logic.

## 4. Code Quality & Clean Code

### ⚠️ Issues to Address

1.  **SRP Violation in `Create`**
    - The `Create` function is excessively long (~300 lines).
    - **Refactoring Plan**:
      - `validateEventRequest(ctx, req)`
      - `generateIdentifiers(ctx, req)`
      - `buildEventModel(ctx, req)`
      - `persistEvent(ctx, event)`

2.  **Magic Numbers**
    - Retry logic in code generation uses hardcoded math: `time.Duration(10 * (attempt + 1))`.
    - **Fix**: Extract `const CodeGenerationBackoffBase = 10 * time.Millisecond`.

3.  **Wide Events Logging**
    - **Implemented**: Moved from scattered `logger.Info` calls to **Context Enrichment**.
    - **Pattern**: Accumulate fields (`event_id`, `step`, `error`) in the context and emit **one** canonical log line at the end of the request.

## 5. Next Steps Checklist

- [ ] **Fix Import Cycle**: Remove `errorc` from `event_model.go`.
- [ ] **Implement Logic**: Add occurrence generation in `event_usecase.go`.
- [ ] **Refactor Create**: Break down the giant function into testable sub-functions.
- [ ] **Add Tests**: Unit tests specifically for the new helper methods (`IsPublic`, `Duration`, etc.).
