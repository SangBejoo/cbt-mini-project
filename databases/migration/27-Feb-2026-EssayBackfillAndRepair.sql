-- Migration: Data repair for essay grading lifecycle
-- Date: 27-Feb-2026

-- Mark completed sessions with essay questions as grading_in_progress if any essay answer has no score
WITH essay_sessions AS (
    SELECT DISTINCT ts.id
    FROM test_session ts
    JOIN test_session_soal tss ON tss.id_test_session = ts.id
    JOIN soal s ON s.id = tss.id_soal
    LEFT JOIN jawaban_siswa js ON js.id_test_session_soal = tss.id
    WHERE s.question_type = 'essay'
      AND ts.status IN ('completed', 'graded', 'grading_in_progress')
      AND (js.id IS NULL OR js.nilai_essay IS NULL)
)
UPDATE test_session ts
SET status = 'grading_in_progress'
WHERE ts.id IN (SELECT id FROM essay_sessions)
  AND ts.status <> 'grading_in_progress';

-- Mark grading_in_progress as graded when all essay answers are scored
WITH fully_graded AS (
    SELECT ts.id
    FROM test_session ts
    JOIN test_session_soal tss ON tss.id_test_session = ts.id
    JOIN soal s ON s.id = tss.id_soal
    LEFT JOIN jawaban_siswa js ON js.id_test_session_soal = tss.id
    WHERE s.question_type = 'essay'
      AND ts.status = 'grading_in_progress'
    GROUP BY ts.id
    HAVING COUNT(*) = COUNT(js.id)
       AND COUNT(*) = COUNT(CASE WHEN js.nilai_essay IS NOT NULL THEN 1 END)
)
UPDATE test_session ts
SET status = 'graded'
WHERE ts.id IN (SELECT id FROM fully_graded);
