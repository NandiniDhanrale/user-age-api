package service

import (
	"context"
	"errors"
	"time"

	"github.com/NandiniDhanrale/user-age-api/db/sqlc"
	"github.com/NandiniDhanrale/user-age-api/internal/models"
	"github.com/NandiniDhanrale/user-age-api/internal/repository"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidDOB        = errors.New("invalid date of birth format, expected YYYY-MM-DD")
	ErrServiceUnavailable = errors.New("database is unavailable")
)

type UserService interface {
	Create(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error)
	GetByID(ctx context.Context, id int32) (models.UserResponse, error)
	List(ctx context.Context) ([]models.UserResponse, error)
	Update(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error)
	Delete(ctx context.Context, id int32) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) checkDB() error {
	if s.repo == nil {
		return ErrServiceUnavailable
	}
	return nil
}

func (s *userService) Create(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	if err := s.checkDB(); err != nil {
		return models.UserResponse{}, err
	}

	dob, err := parseDOB(req.DOB)
	if err != nil {
		return models.UserResponse{}, ErrInvalidDOB
	}

	user, err := s.repo.Create(ctx, req.Name, dob)
	if err != nil {
		return models.UserResponse{}, err
	}

	return toResponse(user), nil
}

func (s *userService) GetByID(ctx context.Context, id int32) (models.UserResponse, error) {
	if err := s.checkDB(); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}
	return toResponse(user), nil
}

func (s *userService) List(ctx context.Context) ([]models.UserResponse, error) {
	if err := s.checkDB(); err != nil {
		return nil, err
	}

	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]models.UserResponse, len(users))
	for i, u := range users {
		responses[i] = toResponse(u)
	}
	return responses, nil
}

func (s *userService) Update(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error) {
	if err := s.checkDB(); err != nil {
		return models.UserResponse{}, err
	}

	dob, err := parseDOB(req.DOB)
	if err != nil {
		return models.UserResponse{}, ErrInvalidDOB
	}

	user, err := s.repo.Update(ctx, id, req.Name, dob)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}
	return toResponse(user), nil
}

func (s *userService) Delete(ctx context.Context, id int32) error {
	if err := s.checkDB(); err != nil {
		return err
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func parseDOB(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func calculateAge(dob time.Time) int {
	now := time.Now()
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	return age
}

func toResponse(user sqlc.User) models.UserResponse {
	return models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		DOB:  user.DOB.Format("2006-01-02"),
		Age:  calculateAge(user.DOB),
	}
}
