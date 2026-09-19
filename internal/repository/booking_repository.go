package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
)

type BookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{db: db}
}

// expose underlying pool so service layer can start transactions
func (r *BookingRepository) Pool() *pgxpool.Pool {
	return r.db
}

func (r *BookingRepository) CreateTx(ctx context.Context, tx pgxTx, b *model.Booking) error {
	query := `
	INSERT INTO bookings
		(reference, customer_name, customer_phone, start_date, end_date, status, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at`

	err := tx.QueryRow(ctx, query, b.Reference, b.CustomerName, b.CustomerPhone, b.StartDate, b.EndDate, b.Status, b.ExpiresAt).Scan(&b.ID, &b.CreatedAt)

	if err != nil {
		log.Fatalln("CreateTx error: failed to insert into bookings")
		return err
	}

	for i := range b.Items {
		item := b.Items[i]
		item.BookingID = b.ID
		_, err := tx.Exec(ctx, `
		INSERT INTO booking_items (booking_id, equipment_item_id, quantity) VALUES($1, $2, $3)`, item.BookingID, item.EquipmentItemID, item.Quantity)

		if err != nil {
			log.Fatalf("CreateTx error: failed to insert into items ID: %d\n", item.ID)
			return err
		}
	}

	return nil
}
