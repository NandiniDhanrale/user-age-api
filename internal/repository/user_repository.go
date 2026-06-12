package repository

import (
	"context"
	"time"

	"github.com/NandiniDhanrale/user-age-api/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error)
	GetByID(ctx context.Context, id int32) (sqlc.User, error)
	List(ctx context.Context) ([]sqlc.User, error)
	Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error)
	Delete(ctx context.Context, id int32) error
}

type userRepository struct {
	q *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{q: sqlc.New(pool)}
}

func (r *userRepository) Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error) {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{Name: name, DOB: dob})
}

func (r *userRepository) GetByID(ctx context.Context, id int32) (sqlc.User, error) {
	return r.q.GetUser(ctx, id)
}

func (r *userRepository) List(ctx context.Context) ([]sqlc.User, error) {
	return r.q.ListUsers(ctx)
}

func (r *userRepository) Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error) {
	return r.q.UpdateUser(ctx, sqlc.UpdateUserParams{ID: id, Name: name, DOB: dob})
}

func (r *userRepository) Delete(ctx context.Context, id int32) error {
	return r.q.DeleteUser(ctx, id)
}
