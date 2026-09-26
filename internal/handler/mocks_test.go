package handler

import (
	"context"
	"time"

	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
)

// Each mock uses a func field per method, so each test can define only
// the behavior it cares about inline, e.g.:
//
//	svc := &mockEquipmentService{
//	    listAvailableFn: func(ctx context.Context, start, end time.Time, categoryID *int) ([]model.EquipmentAvailability, error) {
//	        return []model.EquipmentAvailability{{...}}, nil
//	    },
//	}

type mockEquipmentService struct {
	listAvailableFn func(ctx context.Context, start, end time.Time, categoryID *int) ([]model.EquipmentAvailability, error)
	getBySlugFn     func(ctx context.Context, slug string) (*model.EquipmentItem, error)
}

func (m *mockEquipmentService) ListAvailable(ctx context.Context, start, end time.Time, categoryID *int) ([]model.EquipmentAvailability, error) {
	return m.listAvailableFn(ctx, start, end, categoryID)
}

func (m *mockEquipmentService) GetBySlug(ctx context.Context, slug string) (*model.EquipmentItem, error) {
	return m.getBySlugFn(ctx, slug)
}

type mockBookingService struct {
	checkoutFn                func(ctx context.Context, in service.CheckoutInput) (*model.Booking, string, error)
	lookupByReferenceAndPhone func(ctx context.Context, reference, phone string) (*model.Booking, error)
}

func (m *mockBookingService) Checkout(ctx context.Context, in service.CheckoutInput) (*model.Booking, string, error) {
	return m.checkoutFn(ctx, in)
}

func (m *mockBookingService) LookupByReferenceAndPhone(ctx context.Context, reference, phone string) (*model.Booking, error) {
	return m.lookupByReferenceAndPhone(ctx, reference, phone)
}

type mockAdminBookingService struct {
	getFilteredListFn           func(ctx context.Context, status, reference, phone string, page, limit int) (service.PaginatedBookingResponse, error)
	lookupByReferenceAndPhoneFn func(ctx context.Context, reference, phone string) (*model.Booking, error)
	confirmFn                   func(ctx context.Context, bookingId int) (*model.Booking, error)
	cancelFn                    func(ctx context.Context, bookingId int, reason *model.CancelReason) (*model.Booking, error)
	markPickedUpFn              func(ctx context.Context, bookingId int) (*model.Booking, error)
	markReturnedFn              func(ctx context.Context, bookingId int) (*model.Booking, error)
}

func (m *mockAdminBookingService) GetFilteredList(ctx context.Context, status, reference, phone string, page, limit int) (service.PaginatedBookingResponse, error) {
	return m.getFilteredListFn(ctx, status, reference, phone, page, limit)
}

func (m *mockAdminBookingService) LookupByReferenceAndPhone(ctx context.Context, reference, phone string) (*model.Booking, error) {
	return m.lookupByReferenceAndPhoneFn(ctx, reference, phone)
}

func (m *mockAdminBookingService) Confirm(ctx context.Context, bookingID int) (*model.Booking, error) {
	return m.confirmFn(ctx, bookingID)
}

func (m *mockAdminBookingService) Cancel(ctx context.Context, bookingID int, reason *model.CancelReason) (*model.Booking, error) {
	return m.cancelFn(ctx, bookingID, reason)
}

func (m *mockAdminBookingService) PickedUp(ctx context.Context, bookingID int) (*model.Booking, error) {
	return m.markPickedUpFn(ctx, bookingID)
}

func (m *mockAdminBookingService) Returned(ctx context.Context, bookingID int) (*model.Booking, error) {
	return m.markReturnedFn(ctx, bookingID)
}

type mockAuthService struct {
	loginFn func(ctx context.Context, email, password string) (string, *model.Employee, error)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (string, *model.Employee, error) {
	return m.loginFn(ctx, email, password)
}

type mockEmployeeGetter struct {
	getByIDFn func(ctx context.Context, id int) (*model.Employee, error)
}

func (m *mockEmployeeGetter) GetByID(ctx context.Context, id int) (*model.Employee, error) {
	return m.getByIDFn(ctx, id)
}
