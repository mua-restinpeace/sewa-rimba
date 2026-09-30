package repository

import (
	"context"
	"errors"
	"log"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, name string) (*model.Category, error) {
	var c model.Category
	generatedSlug := slug.Make(name)

	query := `
	INSERT INTO categories(name, slug)
	VALUES($1, $2)
	RETURNING id, name, slug`

	err := r.db.QueryRow(ctx, query, name, generatedSlug).Scan(
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
		log.Printf("GetList error: %s\n", err)
		return nil, err
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

	return categories, rows.Err()
}

func (r *CategoryRepository) Update(ctx context.Context, categoryId int, name string) (*model.Category, error) {
	query := `
	UPDATE categories SET name = $1, slug = $2 where id = $3
	RETURNING id, name, slug`

	var c model.Category
	generatedSlug := slug.Make(name)
	err := r.db.QueryRow(ctx, query, name, generatedSlug, categoryId).Scan(&c.ID, &c.Name, &c.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		log.Printf("Update error: $%s\n", err)
		return nil, err
	}

	return &c, err
}

func (r *CategoryRepository) Delete(ctx context.Context, categoryId int) error {
	query := `
	DELETE FROM categories where id = $1`

	tag, err := r.db.Exec(ctx, query, categoryId)
	if err != nil {
		log.Printf("delete category error: %s\n", err)
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
