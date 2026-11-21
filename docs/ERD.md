# ERD POS Apotek

```mermaid
erDiagram
    USERS ||--o{ PENERIMAAN : mencatat
    USERS ||--o{ PENJUALAN : melayani
    USERS ||--o{ RETUR : menerima
    USERS ||--o{ OPNAME : menghitung
    SATUAN ||--o{ OBAT : satuan_dasar
    OBAT ||--o{ OBAT_SATUAN : memiliki
    SATUAN ||--o{ OBAT_SATUAN : konversi
    PBF ||--o{ PENERIMAAN : memasok
    PENERIMAAN ||--|{ PENERIMAAN_ITEM : berisi
    OBAT_SATUAN ||--o{ PENERIMAAN_ITEM : diterima_dalam
    PENERIMAAN_ITEM ||--|| BATCH : menghasilkan
    OBAT ||--o{ BATCH : memiliki
    PENJUALAN ||--|{ PENJUALAN_ITEM : berisi
    OBAT ||--o{ PENJUALAN_ITEM : dijual
    SATUAN ||--o{ PENJUALAN_ITEM : satuan_jual
    PENJUALAN_ITEM ||--|{ PENJUALAN_ITEM_BATCH : dialokasikan
    BATCH ||--o{ PENJUALAN_ITEM_BATCH : sumber_stok
    PENJUALAN ||--o{ RETUR : diretur
    RETUR ||--|{ RETUR_ITEM : berisi
    PENJUALAN_ITEM_BATCH ||--o{ RETUR_ITEM : batch_asal
    OPNAME ||--|{ OPNAME_ITEM : berisi
    BATCH ||--o{ OPNAME_ITEM : dihitung
```

`penjualan_item_batch` penting karena satu permintaan, misalnya 15 tablet, dapat dipenuhi oleh 5 tablet dari batch A dan 10 tablet dari batch B. HPP dan retur tetap dapat ditelusuri ke batch tepat.
