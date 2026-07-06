---
tags: [tech-debt]
sources: [raw/2026-07-06-bootstrap-go-backend-models.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Duplicated Availability Status Logic

`models.DefineAvailabilityStatus` computes one of `available`/`unavailable`/`full`/`soon`/`walkin` for an event or instance, based on seat counts and the registration time window. [internal/models/event_model.go:491-582](internal/models/event_model.go#L491-L582)

It takes an `interface{}` and type-switches over 4+ concrete DB-output structs (`GetAllEventsDBOutput`, `GetEventByCodeDBOutput`, `GetInstanceByEventCodeDBOutput`, `GetInstanceByCodeDBOutput`) that all expose near-identical fields (`TotalRemainingSeats`, `InstanceTotalSeats`, register start/end times) under different struct names — each branch repeats the same field-extraction logic verbatim.

A shared interface with accessor methods (e.g. `RemainingSeats() int`, `RegisterWindow() (time.Time, time.Time)`) implemented by each DB-output struct could collapse the type-switch into one code path. Worth considering if a fifth DB-output shape is ever added, since the duplication would grow with it.
