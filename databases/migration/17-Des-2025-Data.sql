CREATE TABLE `mata_pelajaran` (
                                  `id` int PRIMARY KEY AUTO_INCREMENT,
                                  `nama` varchar(50) UNIQUE NOT NULL
);

CREATE TABLE `tingkat` (
                           `id` int PRIMARY KEY AUTO_INCREMENT,
                           `nama` varchar(50) UNIQUE NOT NULL
);

CREATE TABLE `materi` (
                          `id` int PRIMARY KEY AUTO_INCREMENT,
                          `id_mata_pelajaran` int NOT NULL,
                          `id_tingkat` int NOT NULL,
                          `nama` varchar(100) NOT NULL
);

CREATE TABLE `soal` (
                        `id` int PRIMARY KEY AUTO_INCREMENT,
                        `id_materi` int NOT NULL,
                        `id_tingkat` int NOT NULL,
                        `pertanyaan` text NOT NULL,
                        `opsi_a` varchar(255) NOT NULL,
                        `opsi_b` varchar(255) NOT NULL,
                        `opsi_c` varchar(255) NOT NULL,
                        `opsi_d` varchar(255) NOT NULL,
                        `jawaban_benar` char(1) NOT NULL
);

CREATE TABLE `test_session` (
                                `id` int PRIMARY KEY AUTO_INCREMENT,
                                `session_token` varchar(64) UNIQUE NOT NULL,
                                `nama_peserta` varchar(100) NOT NULL,
                                `id_tingkat` int NOT NULL,
                                `id_mata_pelajaran` int NOT NULL,
                                `waktu_mulai` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                `waktu_selesai` timestamp NULL,
                                `durasi_menit` int NOT NULL,
                                `nilai_akhir` decimal(5,2) NULL,
                                `jumlah_benar` int NULL,
                                `total_soal` int NULL,
                                `status` enum('ongoing','completed','timeout') DEFAULT 'ongoing'
);

CREATE TABLE `test_session_soal` (
                                     `id` int PRIMARY KEY AUTO_INCREMENT,
                                     `id_test_session` int NOT NULL,
                                     `id_soal` int NOT NULL,
                                     `nomor_urut` int NOT NULL
);

CREATE TABLE `jawaban_siswa` (
                                 `id` int PRIMARY KEY AUTO_INCREMENT,
                                 `id_test_session_soal` int NOT NULL,
                                 `jawaban_dipilih` char(1) NULL,
                                 `is_correct` boolean NOT NULL,
                                 `dijawab_pada` timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX `materi_index_0` ON `materi` (`id_mata_pelajaran`, `id_tingkat`, `nama`);

CREATE INDEX `soal_index_1` ON `soal` (`id_materi`);

CREATE INDEX `soal_index_2` ON `soal` (`id_tingkat`);

CREATE UNIQUE INDEX `test_session_index_2` ON `test_session` (`session_token`);

CREATE INDEX `test_session_index_3` ON `test_session` (`id_tingkat`);

CREATE INDEX `test_session_index_4` ON `test_session` (`id_mata_pelajaran`);

CREATE INDEX `test_session_index_5` ON `test_session` (`waktu_mulai`);

CREATE UNIQUE INDEX `test_session_soal_index_6` ON `test_session_soal` (`id_test_session`, `nomor_urut`);

CREATE INDEX `test_session_soal_index_7` ON `test_session_soal` (`id_test_session`);

CREATE UNIQUE INDEX `jawaban_siswa_index_8` ON `jawaban_siswa` (`id_test_session_soal`);

ALTER TABLE `materi` ADD FOREIGN KEY (`id_mata_pelajaran`) REFERENCES `mata_pelajaran` (`id`);

ALTER TABLE `materi` ADD FOREIGN KEY (`id_tingkat`) REFERENCES `tingkat` (`id`);

ALTER TABLE `soal` ADD FOREIGN KEY (`id_materi`) REFERENCES `materi` (`id`);

ALTER TABLE `soal` ADD FOREIGN KEY (`id_tingkat`) REFERENCES `tingkat` (`id`);

ALTER TABLE `test_session` ADD FOREIGN KEY (`id_tingkat`) REFERENCES `tingkat` (`id`);

ALTER TABLE `test_session` ADD FOREIGN KEY (`id_mata_pelajaran`) REFERENCES `mata_pelajaran` (`id`);

ALTER TABLE `test_session_soal` ADD FOREIGN KEY (`id_test_session`) REFERENCES `test_session` (`id`) ON DELETE CASCADE;

ALTER TABLE `test_session_soal` ADD FOREIGN KEY (`id_soal`) REFERENCES `soal` (`id`);

ALTER TABLE `jawaban_siswa` ADD FOREIGN KEY (`id_test_session_soal`) REFERENCES `test_session_soal` (`id`) ON DELETE CASCADE;

