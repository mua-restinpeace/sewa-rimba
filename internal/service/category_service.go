package service

import (
	"context"

	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, name string) (*model.Category, error){
	return s.repo.Create(ctx, name)
}

func (s *CategoryService) GetBySlug(ctx context.Context, slug string) (*model.Category, error){
	return s.repo.GetBySlug(ctx, slug)
}

func (s *CategoryService) GetList(ctx context.Context) ([]model.Category, error){
	return  s.repo.GetList(ctx)
}
