package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
)

type EmployeeRepository struct {
	db *pgxpool.Pool
}

func NewEmployeeRepository(db *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) GetByEmail(ctx context.Context, email string) (*model.Employee, error) {
	var e model.Employee
	query := `
	SELECT id, name, phone, email, password_hash, created_at
	FROM employees WHERE email = $1
	`
	err := r.db.QueryRow(ctx, query, email).Scan(
		&e.ID, &e.Name, &e.Phone, &e.Email, &e.PasswordHash, &e.CreatedAt,
	)
	if err != nil {
		fmt.Println("GetByEmail error: ", err)
		return nil, err
	}
	return &e, err
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id int) (*model.Employee, error){
	var e model.Employee
	query := `
	SELECT id, name, phone, email, password_hash, created_at
	FROM employees WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.Name, &e.Phone, &e.Email, &e.PasswordHash, &e.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &e, nil
}