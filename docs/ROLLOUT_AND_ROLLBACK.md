# Rollout and Rollback Runbook

## Rollout Order
1. Apply DB migrations:
   - `27-Feb-2026-MateriLMSLinkage.sql`
   - `27-Feb-2026-EssaySupport.sql`
   - `27-Feb-2026-EssayBackfillAndRepair.sql`
2. Deploy CBT backend with updated proto + handlers + repository.
3. Run smoke tests:
   - `GET /v1/auth/profile`
   - Materi create/update with LMS linkage IDs
   - Session complete (non-essay and essay)
   - Grade essay endpoint
4. Monitor logs/metrics for:
   - grpc unauthenticated/permission errors
   - DB enum/column errors
   - grading status anomalies

## Rollback Strategy
- App rollback: deploy previous CBT image immediately.
- DB rollback: avoid destructive rollback; schema is additive.
- Behavior rollback:
   - Disable essay grading endpoint at gateway if needed.
   - Keep old completion path operational for non-essay sessions.

## Post-Rollback Notes
- Added columns/enums remain and are safe for old binary.
- Re-run smoke tests against old binary to confirm compatibility.
