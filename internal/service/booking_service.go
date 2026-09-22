package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
)

var (
	ErrItemUnavailable = errors.New("one or more items are no longer available for the selected dates")
	ErrInvalidDates    = errors.New("end date must be on or after start date, and start date cannot be in the past")
	ErrInvalidStatus   = errors.New("booking is not in a state that allows this action")
)

type CheckoutItem struct {
	EquipmentItemID int
	Quantity        int
}

type CheckoutInput struct {
	CustomerName  string
	CustomerPhone string
	StartDate     time.Time
	EndDate       time.Time
	Items         []CheckoutItem
}

type BookingService struct {
	bookingRepo   *repository.BookingRepository
	equipmentRepo *repository.EquipmentRepository
	holdMinutes   int
	shopWhatsApp  string
}

func NewBookingService(bookingRepo *repository.BookingRepository, equipmentRepo *repository.EquipmentRepository, holdMinutes int, shopWhatsApp string) *BookingService {
	return &BookingService{bookingRepo: bookingRepo, equipmentRepo: equipmentRepo, holdMinutes: holdMinutes, shopWhatsApp: shopWhatsApp}
}

func (s *BookingService) Checkout(ctx context.Context, in CheckoutInput) (*model.Booking, string, error) {
	today := time.Now().Truncate(24 * time.Hour)
	if in.EndDate.Before(in.StartDate) || in.StartDate.Before(today) {
		return nil, "", ErrInvalidDates
	}

	tx, err := s.bookingRepo.Pool().Begin(ctx)
	if err != nil {
		log.Fatalf("Checkout error: %s", err)
	}

	defer tx.Rollback(ctx)

	for _, item := range in.Items {
		available, err := s.equipmentRepo.AvailableQuantityTx(ctx, tx, item.EquipmentItemID, in.StartDate, in.EndDate)
		if err != nil {
			log.Fatalf("Checkout error: %s", err)
			return nil, "", ErrItemUnavailable
		}

		if available < item.Quantity {
			log.Fatalln("Checkout error: item unavailable")
			return nil, "", ErrItemUnavailable
		}
	}

	reference, err := generateReference()
	if err != nil {
		log.Fatalln("Checkout error: failed to generate reference")
		return nil, "", err
	}

	booking := &model.Booking{
		Reference:     reference,
		CustomerName:  in.CustomerName,
		CustomerPhone: in.CustomerPhone,
		StartDate:     in.StartDate,
		EndDate:       in.EndDate,
		Status:        model.StatusPending,
		ExpiresAt:     time.Now().Add(time.Duration(s.holdMinutes) * time.Minute),
	}

	for _, item := range in.Items {
		booking.Items = append(booking.Items, model.BookingItems{
			EquipmentItemID: item.EquipmentItemID,
			Quantity:        item.Quantity,
		})
	}

	if err := s.bookingRepo.CreateTx(ctx, tx, booking); err != nil {
		log.Fatalf("Checkout error: %s\n", err)
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Checkout error: %s\n", err)
		return nil, "", err
	}

	return booking, s.buildWhatsAppURL(booking), nil
}

func (s *BookingService) Confirm(ctx context.Context, bookingId int) (*model.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingId)
	if err != nil {
		log.Fatalf("Confirm error: %s\n", err)
		return nil, err
	}

	if !model.CanTransistion(booking.Status, model.StatusConfirmed) {
		return nil, ErrInvalidStatus
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingId, model.StatusConfirmed, nil); err != nil {
		log.Fatalf("Confirm error: failed to update status\n%s\n", err)
		return nil, err
	}
	return s.bookingRepo.GetByID(ctx, bookingId)
}

func (s *BookingService) PickedUp(ctx context.Context, bookingId int) (*model.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingId)
	if err != nil {
		log.Fatalf("Picked up error: %s\n", err)
		return nil, err
	}

	if !model.CanTransistion(booking.Status, model.StatusOngoing) {
		return nil, ErrInvalidStatus
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingId, model.StatusOngoing, nil); err != nil {
		log.Fatalf("Picked up error: failed to update status to ongoing\n%s\n", err)
		return nil, err
	}

	return s.bookingRepo.GetByID(ctx, bookingId)
}

func (s *BookingService) Returned(ctx context.Context, bookingID int) (*model.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		log.Fatalf("Returned error: %s\n", err)
		return nil, err
	}

	if !model.CanTransistion(booking.Status, model.StatusReturned) {
		return nil, ErrInvalidStatus
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, model.StatusReturned, nil); err != nil {
		log.Fatalf("Returned error: failed to update status\n%s\n", err)
		return nil, err
	}

	return s.bookingRepo.GetByID(ctx, bookingID)
}

func (s *BookingService) Cancel(ctx context.Context, bookingID int, reason *model.CancelReason) (*model.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		log.Fatalf("Cancel error: %s", err)
		return nil, err
	}

	if !model.CanTransistion(booking.Status, model.StatusCancelled) {
		return nil, ErrInvalidStatus
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, model.StatusCancelled, reason); err != nil {
		log.Fatalf("Return error: failed to update status\n%s\n", err)
		return nil, err
	}

	return s.bookingRepo.GetByID(ctx, bookingID)
}

func generateReference() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("BK-%X", b), nil
}

func (s *BookingService) buildWhatsAppURL(b *model.Booking) string {
	text := fmt.Sprintf("Booking %s\nName: %s\nDates: %s to %s\n(see items in app)", b.Reference, b.CustomerName, b.StartDate.Format("2006-01-02"), b.EndDate.Format("2006-01-02"))
	return "https://wa.me/" + s.shopWhatsApp + "?text=" + url.QueryEscape(text)
}
