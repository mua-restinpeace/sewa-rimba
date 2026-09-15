package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthService struct {
	repo      *repository.EmployeeRepository
	jwtSecret []byte
}

func NewAuthService(repo *repository.EmployeeRepository, jwtSecrete string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: []byte(jwtSecrete)}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *model.Employee, error) {
	employee, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		fmt.Println("Login error: ", err)
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(employee.PasswordHash), []byte(password)); err != nil {
		fmt.Println("Login error: ", err)
		return "", nil, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"employee_id": employee.ID,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
	})

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		fmt.Println("Login error: ", err)
		return "", nil, err
	}

	return signed, employee, nil
}

// validate the bearer token and extract the employee ID
func (s *AuthService) ParseToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	idFloat, ok := claims["employee_id"].(float64)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return int(idFloat), nil
}
