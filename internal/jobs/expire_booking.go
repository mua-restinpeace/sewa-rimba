package jobs

import (
	"context"
	"log"
	"time"

	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
)

func StartBookingExpireJob(ctx context.Context, repo *repository.BookingRepository, interval time.Duration){
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select{
		case <- ctx.Done():
			return
		case <-ticker.C:
			count, err := repo.ExpiredPendingBookings(ctx)
			if err != nil {
				log.Printf("expire_bookings: error: %v", err)
				continue
			}
			if count > 0{
				log.Printf("expire_bookings: expired %d pending booking(s)", count)
			}
		}
	}
}