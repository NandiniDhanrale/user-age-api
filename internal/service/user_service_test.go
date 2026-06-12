package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NandiniDhanrale/user-age-api/db/sqlc"
	"github.com/NandiniDhanrale/user-age-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error) {
	args := m.Called(ctx, name, dob)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int32) (sqlc.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *mockUserRepository) List(ctx context.Context) ([]sqlc.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]sqlc.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error) {
	args := m.Called(ctx, id, name, dob)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *mockUserRepository) Delete(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCalculateAge(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		dob  time.Time
		want int
	}{
		{
			name: "birthday passed this year",
			dob:  time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			want: now.Year() - 1990,
		},
		{
			name: "birthday later this year",
			dob:  time.Date(1990, time.December, 31, 0, 0, 0, 0, time.UTC),
			want: now.Year() - 1990 - 1,
		},
		{
			name: "birthday today",
			dob:  time.Date(1990, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			want: now.Year() - 1990,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateAge(tt.dob)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreate_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	dob := time.Date(1990, time.May, 15, 0, 0, 0, 0, time.UTC)
	mockRepo.On("Create", mock.Anything, "Alice", dob).
		Return(sqlc.User{ID: 1, Name: "Alice", DOB: dob}, nil)

	resp, err := svc.Create(context.Background(), models.CreateUserRequest{
		Name: "Alice",
		DOB:  "1990-05-15",
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(1), resp.ID)
	assert.Equal(t, "Alice", resp.Name)
	assert.Equal(t, "1990-05-15", resp.DOB)
	assert.Greater(t, resp.Age, 0)
	mockRepo.AssertExpectations(t)
}

func TestCreate_InvalidDOB(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	_, err := svc.Create(context.Background(), models.CreateUserRequest{
		Name: "Alice",
		DOB:  "not-a-date",
	})

	assert.ErrorIs(t, err, ErrInvalidDOB)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestGetByID_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	dob := time.Date(1990, time.May, 15, 0, 0, 0, 0, time.UTC)
	mockRepo.On("GetByID", mock.Anything, int32(1)).
		Return(sqlc.User{ID: 1, Name: "Alice", DOB: dob}, nil)

	resp, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, int32(1), resp.ID)
	assert.Equal(t, "Alice", resp.Name)
}

func TestGetByID_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, int32(99)).
		Return(sqlc.User{}, errors.New("no rows in result set"))

	_, err := svc.GetByID(context.Background(), 99)

	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestList_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	dob1 := time.Date(1990, time.May, 15, 0, 0, 0, 0, time.UTC)
	dob2 := time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC)
	mockRepo.On("List", mock.Anything).
		Return([]sqlc.User{
			{ID: 1, Name: "Alice", DOB: dob1},
			{ID: 2, Name: "Bob", DOB: dob2},
		}, nil)

	resp, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "Alice", resp[0].Name)
	assert.Equal(t, "Bob", resp[1].Name)
}

func TestList_Empty(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	mockRepo.On("List", mock.Anything).Return([]sqlc.User{}, nil)

	resp, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, resp)
}

func TestUpdate_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	dob := time.Date(1991, time.June, 20, 0, 0, 0, 0, time.UTC)
	mockRepo.On("Update", mock.Anything, int32(1), "Alice Updated", dob).
		Return(sqlc.User{ID: 1, Name: "Alice Updated", DOB: dob}, nil)

	resp, err := svc.Update(context.Background(), 1, models.UpdateUserRequest{
		Name: "Alice Updated",
		DOB:  "1991-06-20",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Alice Updated", resp.Name)
	assert.Equal(t, "1991-06-20", resp.DOB)
}

func TestUpdate_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	dob := time.Date(1991, time.June, 20, 0, 0, 0, 0, time.UTC)
	mockRepo.On("Update", mock.Anything, int32(99), "Alice Updated", dob).
		Return(sqlc.User{}, errors.New("no rows in result set"))

	_, err := svc.Update(context.Background(), 99, models.UpdateUserRequest{
		Name: "Alice Updated",
		DOB:  "1991-06-20",
	})

	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUpdate_InvalidDOB(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	_, err := svc.Update(context.Background(), 1, models.UpdateUserRequest{
		Name: "Alice",
		DOB:  "bad-date",
	})

	assert.ErrorIs(t, err, ErrInvalidDOB)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestDelete_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	mockRepo.On("Delete", mock.Anything, int32(1)).Return(nil)

	err := svc.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	svc := NewUserService(mockRepo)

	mockRepo.On("Delete", mock.Anything, int32(99)).
		Return(errors.New("no rows in result set"))

	err := svc.Delete(context.Background(), 99)

	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestParseDOB_Valid(t *testing.T) {
	dob, err := parseDOB("2000-01-02")
	assert.NoError(t, err)
	assert.Equal(t, 2000, dob.Year())
	assert.Equal(t, time.January, dob.Month())
	assert.Equal(t, 2, dob.Day())
}

func TestParseDOB_Invalid(t *testing.T) {
	_, err := parseDOB("01/02/2000")
	assert.Error(t, err)
}
