package repository

import (
	"context"
	"log"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, name string) (*model.Category, error) {
	var c model.Category
	slug := slug.Make(name)

	query := `
	INSERT INTO categories(name, slug)
	VALUES($1, $2)
	RETURNING id, name, slug`

	err := r.db.QueryRow(ctx, query, name, slug).Scan(
		&c.ID, &c.Name, &c.Slug,
	)

	if err != nil {
		log.Printf("Create category errror: %s", err)
		return nil, err
	}

	return &c, err
}
