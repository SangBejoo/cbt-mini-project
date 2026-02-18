# ADR: CBT-LMS Ownership, Question Source, and Canonical Naming

## Status
Accepted - 2026-02-18

## Context
CBT is integrated with LMS for identity, class context, and content linkage.
Current CBT schema already includes integration columns:
- `materials.lms_module_id`
- `materials.lms_book_id`
- `materials.lms_teacher_material_id`
- `materials.lms_class_id`
- `questions.lms_asset_id`

LMS currently exposes `TeacherMaterial` and `TeacherMaterialItem` with generic item types (including `quiz_link`), but does not expose a structured question-bank entity (question stem/options/correct answer) as canonical source.

## Decision 1 — Domain Ownership Rules

### Rule A: Identity & Academic Context Ownership
- LMS is source-of-truth for user identity, class, school, and membership.
- CBT stores cached references only for execution and grading.

### Rule B: Assessment Content Ownership
- CBT is source-of-truth for executable exam structure and questions (`materials`, `questions`, `drag_drop_questions`, `exam_sessions`, `student_answers`).
- LMS keeps assignment and publication context; CBT keeps test delivery data.

### Rule C: Material Ownership Semantics
- Teacher-scoped material:
  - `materials.lms_teacher_material_id` MUST be set.
  - `materials.owner_user_id` SHOULD be set to local CBT user mapped from LMS teacher.
- System/superadmin material:
  - `materials.lms_book_id` or `materials.lms_module_id` SHOULD be set.
  - `materials.owner_user_id` MAY be NULL (system-owned) or explicit admin owner.

### Rule D: Uniqueness and Guardrails
- A material creation path MUST validate exactly one primary linkage intent:
  - teacher flow (`lms_teacher_material_id`), or
  - module flow (`lms_module_id`), or
  - book flow (`lms_book_id`).
- Reject ambiguous payloads where multiple primary linkage ids are set without explicit override policy.

## Decision 2 — Source of Book Questions

### Canonical Source (Current)
- Book/module questions are authored or imported into CBT.
- LMS book/module data is metadata reference and validation target, not question payload source.

### Operational Policy
- For superadmin flow, UI/API must support CSV/Excel/manual authoring in CBT and tag records using `lms_book_id` / `lms_module_id`.
- If LMS later introduces structured question APIs, integration can shift to pull/sync mode via versioned adapter.

### Why
- Current LMS model contains teacher materials and generic content items (`quiz_link`) but not normalized question entities for exam runtime.
- Keeping CBT as question authority avoids partial sync ambiguity and runtime coupling.

## Decision 3 — Canonical Naming Strategy

### Strategy
Use English canonical naming at service contract and documentation boundaries, while keeping backward-compatible persistence during transition.

### Canonical Vocabulary
- `materials` (legacy: materi)
- `questions` (legacy: soal)
- `question_images` (legacy: soal_gambar)
- `drag_drop_questions` (legacy: soal_drag_drop)
- `exam_sessions` (legacy: test_session)
- `student_answers` (legacy: jawaban_siswa)

### Enforcement Rules
- New protobuf fields, DTOs, and external API docs MUST use canonical English names.
- Legacy names may remain in old handlers/repository internals until migration is completed.
- Add translation/mapping only at boundaries; do not mix Indonesian and English names in one API contract.

## Implementation Notes (Step-by-Step)

### Step 1 (Done): Ownership Contract
- Adopt rules above in service-level validation for create/update material flows.
- Add explicit validation errors for ambiguous linkage ids.

### Step 2 (Done): Book Question Source Contract
- Freeze policy: question authoring/import in CBT; LMS only metadata validator.
- Document this in API and integration guides.

### Step 3 (Done): Canonical Naming Contract
- Freeze canonical term set for all new contracts.
- Keep compatibility mapper for legacy field names until migration phase.

## Consequences
- Positive:
  - Clear responsibility split prevents ownership confusion.
  - Essay and mixed-question roadmap remains fully in CBT control.
  - New API evolution remains consistent in English naming.
- Trade-off:
  - Temporary dual naming exists internally until compatibility migration is finished.

## Follow-up (Next TODOs)
- Define compatibility migration plan (non-breaking).
- Update protobuf contracts safely for new/renamed fields.
- Refactor repository mappings and add contract tests.
