SET NAMES utf8mb4;
SET time_zone = '+07:00';

CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('user', 'admin', 'apoteker') NOT NULL DEFAULT 'user',
    aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE satuan (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(30) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE obat (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(120) NOT NULL,
    satuan_dasar_id BIGINT UNSIGNED NOT NULL,
    aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_obat_satuan_dasar FOREIGN KEY (satuan_dasar_id) REFERENCES satuan(id),
    INDEX idx_obat_nama (nama)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE obat_satuan (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    obat_id BIGINT UNSIGNED NOT NULL,
    satuan_id BIGINT UNSIGNED NOT NULL,
    isi INT UNSIGNED NOT NULL COMMENT 'Jumlah satuan dasar',
    harga_jual BIGINT UNSIGNED NOT NULL,
    CONSTRAINT fk_obat_satuan_obat FOREIGN KEY (obat_id) REFERENCES obat(id),
    CONSTRAINT fk_obat_satuan_satuan FOREIGN KEY (satuan_id) REFERENCES satuan(id),
    CONSTRAINT chk_obat_satuan_isi CHECK (isi > 0),
    UNIQUE KEY uk_obat_satuan (obat_id, satuan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pbf (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(120) NOT NULL,
    telepon VARCHAR(30) NULL,
    alamat TEXT NULL,
    aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE penerimaan (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    no_faktur VARCHAR(60) NOT NULL UNIQUE,
    pbf_id BIGINT UNSIGNED NOT NULL,
    tanggal DATE NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    total BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_penerimaan_pbf FOREIGN KEY (pbf_id) REFERENCES pbf(id),
    CONSTRAINT fk_penerimaan_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE penerimaan_item (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    penerimaan_id BIGINT UNSIGNED NOT NULL,
    obat_satuan_id BIGINT UNSIGNED NOT NULL,
    qty INT UNSIGNED NOT NULL,
    harga_beli BIGINT UNSIGNED NOT NULL COMMENT 'Harga beli per satuan penerimaan',
    subtotal BIGINT UNSIGNED NOT NULL,
    CONSTRAINT fk_terima_item_header FOREIGN KEY (penerimaan_id) REFERENCES penerimaan(id),
    CONSTRAINT fk_terima_item_obat_satuan FOREIGN KEY (obat_satuan_id) REFERENCES obat_satuan(id),
    CONSTRAINT chk_terima_item_qty CHECK (qty > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE batch (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    penerimaan_item_id BIGINT UNSIGNED NOT NULL UNIQUE,
    obat_id BIGINT UNSIGNED NOT NULL,
    no_batch VARCHAR(80) NOT NULL,
    expired_at DATE NOT NULL,
    harga_beli_dasar DECIMAL(15,4) UNSIGNED NOT NULL,
    stok_dasar INT UNSIGNED NOT NULL,
    stok_awal_dasar INT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_batch_terima_item FOREIGN KEY (penerimaan_item_id) REFERENCES penerimaan_item(id),
    CONSTRAINT fk_batch_obat FOREIGN KEY (obat_id) REFERENCES obat(id),
    UNIQUE KEY uk_batch_obat_nomor (obat_id, no_batch),
    INDEX idx_batch_fefo (obat_id, expired_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE penjualan (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    kode VARCHAR(30) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    tanggal DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    total BIGINT UNSIGNED NOT NULL,
    bayar BIGINT UNSIGNED NOT NULL,
    kembali BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_penjualan_user FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_penjualan_tanggal (tanggal)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE penjualan_item (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    penjualan_id BIGINT UNSIGNED NOT NULL,
    obat_id BIGINT UNSIGNED NOT NULL,
    satuan_id BIGINT UNSIGNED NOT NULL,
    qty INT UNSIGNED NOT NULL,
    qty_dasar INT UNSIGNED NOT NULL,
    harga_jual BIGINT UNSIGNED NOT NULL COMMENT 'Snapshot harga jual per satuan',
    subtotal BIGINT UNSIGNED NOT NULL,
    CONSTRAINT fk_jual_item_header FOREIGN KEY (penjualan_id) REFERENCES penjualan(id),
    CONSTRAINT fk_jual_item_obat FOREIGN KEY (obat_id) REFERENCES obat(id),
    CONSTRAINT fk_jual_item_satuan FOREIGN KEY (satuan_id) REFERENCES satuan(id),
    CONSTRAINT chk_jual_item_qty CHECK (qty > 0 AND qty_dasar > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE penjualan_item_batch (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    penjualan_item_id BIGINT UNSIGNED NOT NULL,
    batch_id BIGINT UNSIGNED NOT NULL,
    qty_dasar INT UNSIGNED NOT NULL,
    harga_beli_dasar DECIMAL(15,4) UNSIGNED NOT NULL COMMENT 'Snapshot HPP per satuan dasar',
    CONSTRAINT fk_jual_batch_item FOREIGN KEY (penjualan_item_id) REFERENCES penjualan_item(id),
    CONSTRAINT fk_jual_batch_batch FOREIGN KEY (batch_id) REFERENCES batch(id),
    CONSTRAINT chk_jual_batch_qty CHECK (qty_dasar > 0),
    UNIQUE KEY uk_jual_item_batch (penjualan_item_id, batch_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE retur (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    kode VARCHAR(30) NOT NULL UNIQUE,
    penjualan_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    tanggal DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    alasan TEXT NOT NULL,
    total_refund BIGINT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_retur_penjualan FOREIGN KEY (penjualan_id) REFERENCES penjualan(id),
    CONSTRAINT fk_retur_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE retur_item (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    retur_id BIGINT UNSIGNED NOT NULL,
    penjualan_item_batch_id BIGINT UNSIGNED NOT NULL,
    qty_dasar INT UNSIGNED NOT NULL,
    kondisi ENUM('baik', 'rusak') NOT NULL,
    kembali_ke_stok BOOLEAN NOT NULL DEFAULT FALSE,
    nilai_refund BIGINT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_retur_item_header FOREIGN KEY (retur_id) REFERENCES retur(id),
    CONSTRAINT fk_retur_item_alokasi FOREIGN KEY (penjualan_item_batch_id) REFERENCES penjualan_item_batch(id),
    CONSTRAINT chk_retur_item_qty CHECK (qty_dasar > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE opname (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    kode VARCHAR(30) NOT NULL UNIQUE,
    tanggal DATE NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    catatan TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_opname_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE opname_item (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    opname_id BIGINT UNSIGNED NOT NULL,
    batch_id BIGINT UNSIGNED NOT NULL,
    stok_sistem INT UNSIGNED NOT NULL,
    stok_fisik INT UNSIGNED NOT NULL,
    selisih INT NOT NULL,
    alasan VARCHAR(255) NULL,
    CONSTRAINT fk_opname_item_header FOREIGN KEY (opname_id) REFERENCES opname(id),
    CONSTRAINT fk_opname_item_batch FOREIGN KEY (batch_id) REFERENCES batch(id),
    UNIQUE KEY uk_opname_batch (opname_id, batch_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
