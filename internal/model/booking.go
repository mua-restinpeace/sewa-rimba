package model

import "time"

type BookingStatus string

const (
	StatusPending   BookingStatus = "pending"
	StatusConfirmed BookingStatus = "confirmed"
	StatusOngoing   BookingStatus = "ongoing"
	StatusReturned  BookingStatus = "returned"
	StatusCancelled BookingStatus = "cancelled"
	StatusExpired   BookingStatus = "expired"
)

type CancelReason string

const (
	ReasonCustomerRequest CancelReason = "customer_request"
	ReasonNoShow          CancelReason = "no_show"
)

type Booking struct {
	ID            int            `json:"id"`
	Reference     string         `json:"reference"`
	CustomerName  string         `json:"customer_name"`
	CustomerPhone string         `json:"customer_phone"`
	StartDate     time.Time      `json:"start_date"`
	EndDate       time.Time      `json:"end_date"`
	Status        BookingStatus  `json:"status"`
	CancelReason  *CancelReason  `json:"cancel_reason,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
	ConfirmedAt   *time.Time     `json:"confirmed_at,omitempty"`
	PickedAt      *time.Time     `json:"picked_at,omitempty"`
	ReturnedAt    *time.Time     `json:"returned_at,omitempty"`
	CancelledAt   *time.Time     `json:"cancelled_at,omitempty"`
	Items         []BookingItems `json:"items,omitempty"`
}

type BookingItems struct {
	ID              int    `json:"id"`
	BookingID       int    `json:"booking_id"`
	EquipmentItemID int    `json:"equipment_item_id"`
	EquipmentName   string `json:"equipment_name"`
	Quantity        int    `json:"quantity"`
}

var allowedTransitions = map[BookingStatus][]BookingStatus{
	StatusPending:   {StatusConfirmed, StatusCancelled, StatusExpired},
	StatusConfirmed: {StatusOngoing, StatusCancelled},
	StatusOngoing:   {StatusReturned},
}

func CanTransistion(from, to BookingStatus) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}

	return false
}
