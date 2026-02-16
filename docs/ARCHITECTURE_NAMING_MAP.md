# Architecture Naming Map (CBT ↔ LMS)

This document defines canonical cross-service terminology to reduce semantic mismatch.

## Canonical Terms

- **Learning Module**: material/exam package consumed by assignment flow.
- **Class**: LMS class context.
- **Subject**: mata pelajaran.
- **Level**: tingkat.

## Current Physical Names (Legacy-Compatible)

| Canonical | LMS | CBT |
|---|---|---|
| Learning Module | `modules` (class-scoped) | `materi` (subject+level scoped + LMS bridge columns) |
| Subject | `subjects` | `mata_pelajaran` |
| Level | `levels` | `tingkat` |
| Student Answer | `submissions/assignment_attempts` | `jawaban_siswa` |

## Mapping Rules (Current)

1. `module_id` in assignment events is treated as a **module reference**.
2. Resolver priority in CBT sync:
   - first by `materi.lms_module_id`,
   - fallback by local `materi.id`.
3. Preserve backward compatibility while contracts and docs move to canonical vocabulary.

## Non-Goals (for now)

- No destructive table/column rename in CBT or LMS yet.
- No API breaking changes while still in active development.

## Migration Direction

- Keep compatibility resolver during development.
- Move incrementally to canonical DTO naming and eventually to unified DB naming after planned migration.
