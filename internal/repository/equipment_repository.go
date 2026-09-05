package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
)

type EquipmentRepository struct {
	db *pgxpool.Pool
}

func NewEquipmentrRepository(db *pgxpool.Pool) *EquipmentRepository {
	return &EquipmentRepository{db: db}
}

// list available active equipment filtered by category
// with availability computed against [start, end] dates
func (r *EquipmentRepository) ListAvailable(ctx context.Context, start, end time.Time, categoryID *int) ([]model.EquipmentAvailability, error) {
	query := `
	SELECT
		ei.id, ei.category_id, ei.name, ei.slug, COALESCE(ei.description, ''),
		ei.daily_rate, ei.total_quantity, COALESCE(ei.condition_notes, ''), COALESCE(ei.photo_url, ''),
		ei.is_active, ei.created_at, ei.updated_at,
		ei.total_quantity - COALESCE(SUM(bi.quantity), 0) as available_quantity
	FROM equipment_items ei
	LEFT JOIN booking_items bi ON bi.equipment_item_id = ei.id
	LEFT JOIN bookings b on b.id = bi.booking_id
		AND b.status IN ('pending', 'confirmed', 'ongoing')
		AND b.start_date <= $2::date
		AND b.end_date >= $1::date
	WHERE ei.is_active = true
		AND ($3::int IS NULL OR ei.category_id = $3::int)
	GROUP BY ei.id
	ORDER BY ei.name`

	rows, err := r.db.Query(ctx, query, start, end, categoryID)
	if err != nil {
		fmt.Println("query error: ", err)
		return nil, err
	}
	defer rows.Close()

	var items []model.EquipmentAvailability
	for rows.Next() {
		var e model.EquipmentAvailability
		if err := rows.Scan(&e.ID, &e.CategoryID, &e.Name, &e.Slug, &e.Description, &e.DailyRate, &e.TotalQuantity, &e.Condition_Notes, &e.PhotoURL, &e.IsActive, &e.CreatedAt, &e.UpdateAt, &e.AvailableQuantity); err != nil {
			fmt.Println("scan error: ", err)
			return nil, err
		}

		items = append(items, e)
	}

	return items, rows.Err()
}
