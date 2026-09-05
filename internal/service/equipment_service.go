package service

import (
	"context"
	"time"

	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
)

type EquipmentService struct {
	repo *repository.EquipmentRepository
}

func NewEquipmentService(repo *repository.EquipmentRepository) *EquipmentService {
	return &EquipmentService{repo: repo}
}

func (s *EquipmentService) ListAvailable(ctx context.Context, start, end time.Time, categoryID *int) ([]model.EquipmentAvailability, error) {
	return s.repo.ListAvailable(ctx, start, end, categoryID)
}

func (s *EquipmentService) GetBySlug(ctx context.Context, slug string) (*model.EquipmentItem, error){
	return s.repo.GetBySlug(ctx, slug)
}
