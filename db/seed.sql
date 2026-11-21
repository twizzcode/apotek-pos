INSERT INTO users (nama, username, password_hash, role) VALUES
('Administrator', 'admin', '$2a$10$HRYosgXiAJzf22WZaoV9EejbDDsVUsZ9wjEpYOaO2sCJu5oH0shlq', 'admin'),
('Apoteker', 'apoteker', '$2a$10$HRYosgXiAJzf22WZaoV9EejbDDsVUsZ9wjEpYOaO2sCJu5oH0shlq', 'apoteker'),
('Pengguna', 'user', '$2a$10$HRYosgXiAJzf22WZaoV9EejbDDsVUsZ9wjEpYOaO2sCJu5oH0shlq', 'user');

INSERT INTO satuan (nama) VALUES ('Tablet'), ('Kaplet'), ('Strip'), ('Box'), ('Botol'), ('Tube');

INSERT INTO obat (nama, satuan_dasar_id) VALUES
('Paracetamol 500 mg', 1), ('Amoxicillin 500 mg', 2), ('Vitamin C 500 mg', 1),
('OBH Sirup 100 ml', 5), ('Salep Antijamur 10 g', 6);

INSERT INTO obat_satuan (obat_id, satuan_id, isi, harga_jual) VALUES
(1, 1, 1, 1000), (1, 3, 10, 9000), (1, 4, 100, 85000),
(2, 2, 1, 1500), (2, 3, 10, 14000), (2, 4, 100, 130000),
(3, 1, 1, 750), (3, 3, 10, 7000),
(4, 5, 1, 18000), (5, 6, 1, 22000);

INSERT INTO pbf (nama, telepon, alamat) VALUES
('PBF Sehat Sentosa', '021-5550101', 'Jakarta'),
('PBF Medika Utama', '021-5550102', 'Bogor');
