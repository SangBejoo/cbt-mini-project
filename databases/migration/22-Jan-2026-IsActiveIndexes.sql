-- Performance optimization indexes for active status and user limits
CREATE INDEX mata_pelajaran_is_active_index ON mata_pelajaran (is_active);

CREATE INDEX tingkat_is_active_index ON tingkat (is_active);

CREATE INDEX soal_is_active_index ON soal (is_active);

-- Composite index for user_limits lookups (covers WHERE user_id = X AND limit_type = Y)
CREATE INDEX user_limits_user_type_index ON user_limits (user_id, limit_type);

-- Additional indexes for frequently accessed foreign key relationships
CREATE INDEX soal_gambar_id_soal_urutan_index ON soal_gambar (id_soal, urutan);

CREATE INDEX materi_id_index ON materi (id);

-- Index for user_limit_usage inserts and queries
CREATE INDEX user_limit_usage_user_created_index ON user_limit_usage (user_id, created_at);
