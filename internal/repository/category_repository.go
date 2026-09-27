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

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*model.Category, error) {
	query := `
	SELECT id, name, slug
	FROM categories where slug = $1`

	var c model.Category
	err := r.db.QueryRow(ctx, query, slug).Scan(&c.ID, &c.Name, &c.Slug)
	if err != nil {
		log.Printf("GetBySlug error: %s", err)
		return nil, err
	}

	return &c, err
}

func (r *CategoryRepository) GetList(ctx context.Context) ([]model.Category, error) {
	query := `
	SELECT id, name, slug
	FROM categories`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		log.Printf("GetList error: %s\n", rows.Err())
		return nil, rows.Err()
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			log.Printf("GetList scan error: %s\n", err)
			return nil, err
		}

		categories = append(categories, c)
	}

	return categories, err
}
