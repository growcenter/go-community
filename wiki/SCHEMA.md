# Wiki Schema

Conventions for this repo's wiki. Agents and humans both follow this.

## Folder layout

- `README.md` — the catalog. Auto-rewritten on every ingest. Don't hand-edit (your changes will be overwritten).
- `SCHEMA.md` — this file. Hand-edited when team conventions change.
- `log.md` — append-only chronological ingest log. Don't reorder or edit past entries.
- `raw/` — human-contributed source docs. Immutable once committed. The agent reads these but never modifies them.
- `entities/` — agent-maintained concept, component, and decision pages. Filenames are `kebab-case.md`.
- `LINT-REPORT.md` (generated) — output of `/codewiki:lint`. Conventionally not committed.

## Entity page frontmatter

Every file in `entities/` starts with:

```yaml
---
tags: [<category>, ...]
sources: [raw/<file>.md, ...]
updated: YYYY-MM-DD
ingested_at: <short-sha>
---
```

- `tags`: free-form list. Used by the catalog for grouping.
- `sources`: raw docs this page draws from. Maintained by the agent.
- `updated`: ISO date, set by the agent on any write.
- `ingested_at`: short SHA of `HEAD` at the last ingest that touched this page. Used by lint to detect code-citation drift.

No explicit author list — contributors are derived from `git log` over the `sources` paths.

## Recommended tag vocabulary

Non-enforcing. Teams can add their own.

- `architecture` — system components, modules, services.
- `concept` — domain concepts, invariants, business rules.
- `tech-debt` — known debt documented as wiki entries. Surfaces as a dedicated section in the catalog.
- `decision` — trade-off decisions and their context.
- `runbook` — operational procedures.

## Links

Use standard markdown relative links only:

- `[Entity Name](entities/entity-name.md)` within the wiki.
- `[path/to/file:42](path/to/file#L42)` to cite code (single line).
- `[path/to/file:42-58](path/to/file#L42-L58)` to cite a code range.

Do not use `[[wikilinks]]`. GitHub and Bitbucket don't render them.

## Code citations

The audience is developers, so citing the codebase is expected. Whenever an entity page makes a claim derived from code, cite the location in the format above.

Rules:

- Paths relative to the repo root.
- No absolute URLs, no commit-pinned permalinks (they rot faster than line numbers).
- If a claim can't be tied to a specific file and line, rewrite it to be less specific or cite the raw doc instead.

## Log entry format

Append-only. Each ingest adds:

```
## [YYYY-MM-DD] ingest | <doc title> | author: <name> | commit: <short-sha> | at: <ingest-short-sha>

<1–3 sentence prose summary of what this raw doc contributed.>

Touched:
- entities/<page>.md
- entities/<other>.md
```

- `commit:` — short SHA of the commit that first added the raw doc. Omitted if uncommitted.
- `at:` — short SHA of `HEAD` at ingest time.

## Catalog structure (`README.md`)

The agent rewrites `README.md` on every ingest. Layout:

1. **Intro** — 1–3 lines. What this wiki is. How to contribute.
2. **Recently updated** — top 10 entities by `updated:` desc.
3. **Browse by category** — one subsection per distinct `tags:` value, entities alphabetized within.
4. **Raw sources** — alphabetical list of `raw/*` with one-line summary and author.
5. **Health** (optional) — one-line summary of the most recent `LINT-REPORT.md` if generated in the last 30 days.

All four derived sections are deterministic given the filesystem state. The agent must not hand-curate ordering.
