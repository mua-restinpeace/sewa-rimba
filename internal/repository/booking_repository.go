package repository

import (
	"context"
	"log"
	"strconv"
	"time"

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

const bookingSelectQuery = `
	SELECT 	b.id, b.reference, b.customer_name, b.	customer_phone, b.start_date, b.end_date, b.	status, b.cancel_reason, b.created_at, b.expires_at, b.confirmed_at, b.picked_up_at, b.returned_at, b.cancelled_at
	FROM bookings b`

func (r *BookingRepository) scanOne(ctx context.Context, query string, args ...interface{}) (*model.Booking, error) {
	var b model.Booking
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&b.ID, &b.Reference, &b.CustomerName, &b.CustomerPhone, &b.StartDate, &b.EndDate, &b.Status, &b.CancelReason, &b.CreatedAt, &b.ExpiresAt, &b.ConfirmedAt, &b.PickedAt, &b.ReturnedAt, &b.CancelledAt,
	)

	if err != nil {
		log.Fatalf("scanOne eror: %s\n", err)
		return nil, err
	}

	items, err := r.getItems(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	b.Items = items
	return &b, nil
}

func (r *BookingRepository) getItems(ctx context.Context, bookingID int) ([]model.BookingItems, error) {
	query := `SELECT
		bi.id, bi.booking_id, bi.equipment_item_id, bi.quantity, ei.name
		FROM booking_items bi
		JOIN equipment_items ei ON ei.id = bi.equipment_item_id
		WHERE bi.booking_id = $1`

	rows, err := r.db.Query(ctx, query, bookingID)
	if err != nil {
		log.Fatalf("getItems error: %s", err)
		return nil, err
	}
	defer rows.Close()

	var items []model.BookingItems
	for rows.Next() {
		var it model.BookingItems
		if err := rows.Scan(&it.ID, &it.BookingID, &it.EquipmentItemID, &it.Quantity, &it.EquipmentName); err != nil {
			return nil, err
		}

		items = append(items, it)
	}

	return items, rows.Err()
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID int) (*model.Booking, error) {
	query := bookingSelectQuery + ` WHERE b.id = $1`
	return r.scanOne(ctx, query, bookingID)
}

func (r *BookingRepository) GetByReferenceAndPhone(ctx context.Context, reference, phone string)  (*model.Booking, error){
	query := bookingSelectQuery + ` WHERE b.reference = $1 AND b.customer_phone = $2`
	return  r.scanOne(ctx, query, reference, phone)
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, bookingID int, newStatus model.BookingStatus, cancelReason *model.CancelReason) error {
	now := time.Now()
	var timestampCol string
	switch newStatus {
	case model.StatusConfirmed:
		timestampCol = "confirmed_at"
	case model.StatusOngoing:
		timestampCol = "picked_up_at"
	case model.StatusReturned:
		timestampCol = "returned_at"
	case model.StatusCancelled:
		timestampCol = "cancelled_at"
	}

	query := `UPDATE bookings SET status = $1`
	args := []interface{}{newStatus}
	argN := 2

	if timestampCol != "" {
		query += `, ` + timestampCol + ` = $` + strconv.Itoa(argN)
		args = append(args, now)
		argN++
	}

	if cancelReason != nil {
		query += `, cancel_reason = $` + strconv.Itoa(argN)
		args = append(args, *cancelReason)
		argN++
	}

	query += ` WHERE id = $` + strconv.Itoa(argN)
	args = append(args, bookingID)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// ExpiredPendingBokings is called by the background job
func (r *BookingRepository) ExpiredPendingBookings(ctx context.Context,)(int64, error){
	query := `
	UPDATE bookings SET status = 'expired'
	WHERE status = 'pending' AND expires_at < now()`
	
	tag, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}