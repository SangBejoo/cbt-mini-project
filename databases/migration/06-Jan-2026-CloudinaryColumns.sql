-- Add Cloudinary columns to soal_gambar table
-- Date: January 6, 2026

ALTER TABLE soal_gambar ADD COLUMN cloud_id VARCHAR(255) NULL AFTER keterangan;
ALTER TABLE soal_gambar ADD COLUMN public_id VARCHAR(500) NULL AFTER cloud_id;

CREATE INDEX idx_soal_gambar_cloud_id ON soal_gambar(cloud_id);
CREATE INDEX idx_soal_gambar_public_id ON soal_gambar(public_id);