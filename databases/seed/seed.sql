-- Seed data for CBT Mini Project

-- Insert Mata Pelajaran
INSERT INTO mata_pelajaran (nama) VALUES ('Matematika');

-- Insert Tingkat
INSERT INTO tingkat (nama) VALUES ('Tingkat 1'), ('Tingkat 2'), ('Tingkat 3');

-- Insert Materi
INSERT INTO materi (id_mata_pelajaran, id_tingkat, nama) VALUES
(1, 1, 'Aljabar Dasar'),
(1, 1, 'Geometri'),
(1, 2, 'Trigonometri'),
(1, 3, 'Kalkulus');

-- Insert Soal (20 questions)
INSERT INTO soal (id_materi, id_tingkat, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, jawaban_benar) VALUES
(1, 1, 'Berapakah hasil dari 2 + 2?', '3', '4', '5', '6', 'B'),
(1, 1, 'Berapakah hasil dari 5 x 3?', '8', '15', '20', '25', 'B'),
(1, 1, 'Berapakah hasil dari 10 - 4?', '4', '5', '6', '7', 'C'),
(1, 1, 'Berapakah hasil dari 12 / 3?', '2', '3', '4', '5', 'C'),
(1, 1, 'Berapakah hasil dari 7 + 8?', '14', '15', '16', '17', 'B'),
(1, 1, 'Berapakah hasil dari 9 x 2?', '16', '17', '18', '19', 'C'),
(1, 1, 'Berapakah hasil dari 20 - 7?', '11', '12', '13', '14', 'C'),
(1, 1, 'Berapakah hasil dari 6 / 2?', '2', '3', '4', '5', 'B'),
(1, 1, 'Berapakah hasil dari 4 + 9?', '11', '12', '13', '14', 'C'),
(1, 1, 'Berapakah hasil dari 8 x 5?', '35', '40', '45', '50', 'B'),
(2, 1, 'Apakah nama bangun datar dengan 4 sisi sama panjang?', 'Segitiga', 'Persegi', 'Lingkaran', 'Trapesium', 'B'),
(2, 1, 'Berapakah luas persegi dengan sisi 5 cm?', '20 cm²', '25 cm²', '30 cm²', '35 cm²', 'B'),
(2, 1, 'Apakah nama bangun ruang dengan 6 sisi?', 'Kubus', 'Balok', 'Prisma', 'Limas', 'A'),
(2, 1, 'Berapakah keliling persegi dengan sisi 6 cm?', '18 cm', '24 cm', '30 cm', '36 cm', 'B'),
(2, 1, 'Apakah sudut dalam segitiga sama kaki?', '90 derajat', '180 derajat', '360 derajat', 'Tidak tentu', 'D'),
(3, 2, 'Berapakah nilai sin 90°?', '0', '0.5', '1', '√2/2', 'C'),
(3, 2, 'Berapakah nilai cos 0°?', '0', '0.5', '1', '√2/2', 'C'),
(3, 2, 'Apakah rumus Pythagoras?', 'a² + b² = c²', 'a² - b² = c²', 'a² x b² = c²', 'a² / b² = c²', 'A'),
(3, 2, 'Berapakah nilai tan 45°?', '0', '1', '√3', '∞', 'B'),
(4, 3, 'Berapakah turunan dari x²?', 'x', '2x', 'x²', '2', 'B');</content>
<parameter name="filePath">d:\lamar pt penerbit erlangga\cbt_mini_project\cbt_mini_project\databases\seed\seed.sql