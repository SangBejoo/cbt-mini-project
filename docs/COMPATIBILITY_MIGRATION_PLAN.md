# Compatibility Migration Plan (CBT Legacy -> Canonical)

## Objective
Deliver non-breaking migration from legacy runtime tables/fields to canonical contracts while keeping production stable.

## Scope
- Material linkage: `lms_module_id`, `lms_book_id`, `lms_teacher_material_id`
- Essay support: question type, essay answers, grading workflow
- Status lifecycle: `ongoing/completed/timeout/scheduled/grading_in_progress/graded`

## Phases
1. Schema additive only
   - Add columns/enums/indexes using `IF NOT EXISTS`.
   - No destructive rename/drop.
2. Dual-read / dual-write
   - Service writes new columns while preserving old behavior.
   - Handlers/read models expose canonical fields.
3. Contract rollout
   - Add protobuf fields and RPCs using new field numbers only.
   - Keep existing fields untouched for compatibility.
4. Backfill and repair
   - Run data repair migration for essay sessions/status.
5. Enforcement
   - Enable ownership guardrails (single primary LMS linkage).
6. Cleanup (future)
   - Remove dead paths only after full client cutover.

## Safety Rules
- Migrations are idempotent.
- No enum/value reordering.
- No field-number reuse in protobuf.
- Feature flags/gradual deploy preferred for new RPCs.

## Validation Checklist
- Create materi with each linkage mode works.
- Mixed sessions (MC + drag + essay) can be completed.
- Essay grading transitions status correctly.
- Existing non-essay flows unchanged.
