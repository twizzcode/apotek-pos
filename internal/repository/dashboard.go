package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type DashboardRepo struct{ DB *sql.DB }

func (r *DashboardRepo) Ringkasan(ctx context.Context) (model.Dashboard, error) {
	var d model.Dashboard
	err := r.DB.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM obat WHERE aktif = TRUE AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM batch WHERE stok_dasar > 0),
			(SELECT COUNT(*) FROM batch WHERE stok_dasar > 0 AND expired_at BETWEEN CURDATE() AND DATE_ADD(CURDATE(), INTERVAL 90 DAY)),
			(SELECT COALESCE(SUM(total), 0) FROM penjualan WHERE tanggal >= CURDATE() AND tanggal < DATE_ADD(CURDATE(), INTERVAL 1 DAY))`).
		Scan(&d.JumlahObat, &d.JumlahBatch, &d.BatchMendekatiED, &d.PenjualanHariIni)
	if err != nil {
		return d, fmt.Errorf("ringkasan dashboard: %w", err)
	}
	return d, nil
}
