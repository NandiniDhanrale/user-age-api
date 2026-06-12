package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/NandiniDhanrale/user-age-api/internal/models"
	"github.com/NandiniDhanrale/user-age-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Create(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *mockUserService) GetByID(ctx context.Context, id int32) (models.UserResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *mockUserService) List(ctx context.Context) ([]models.UserResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.UserResponse), args.Error(1)
}

func (m *mockUserService) Update(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *mockUserService) Delete(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestApp(mockSvc *mockUserService) *fiber.App {
	app := fiber.New()
	h := NewUserHandler(mockSvc)

	api := app.Group("/api", func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not_found"})
		}
		if errors.Is(err, service.ErrInvalidDOB) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_dob"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal"})
	})

	users := api.Group("/users")
	users.Post("/", h.Create)
	users.Get("/", h.List)
	users.Get("/:id", h.GetByID)
	users.Put("/:id", h.Update)
	users.Delete("/:id", h.Delete)

	return app
}

func TestCreate_Success(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	expected := models.UserResponse{ID: 1, Name: "Alice", DOB: "1990-05-15", Age: 34}
	mockSvc.On("Create", mock.Anything, models.CreateUserRequest{Name: "Alice", DOB: "1990-05-15"}).
		Return(expected, nil)

	req, _ := http.NewRequest("POST", "/api/users",
		strings.NewReader(`{"name":"Alice","dob":"1990-05-15"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var body models.UserResponse
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "Alice", body.Name)
	mockSvc.AssertExpectations(t)
}

func TestCreate_ValidationError(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	req, _ := http.NewRequest("POST", "/api/users",
		strings.NewReader(`{"name":"","dob":"1990-05-15"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestCreate_BadJSON(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	req, _ := http.NewRequest("POST", "/api/users",
		strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestList_Success(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	users := []models.UserResponse{
		{ID: 1, Name: "Alice", DOB: "1990-05-15", Age: 34},
		{ID: 2, Name: "Bob", DOB: "1995-01-01", Age: 29},
	}
	mockSvc.On("List", mock.Anything).Return(users, nil)

	req, _ := http.NewRequest("GET", "/api/users", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body []models.UserResponse
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Len(t, body, 2)
}

func TestList_Empty(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	mockSvc.On("List", mock.Anything).Return([]models.UserResponse{}, nil)

	req, _ := http.NewRequest("GET", "/api/users", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body []models.UserResponse
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Empty(t, body)
}

func TestGetByID_Success(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	expected := models.UserResponse{ID: 1, Name: "Alice", DOB: "1990-05-15", Age: 34}
	mockSvc.On("GetByID", mock.Anything, int32(1)).Return(expected, nil)

	req, _ := http.NewRequest("GET", "/api/users/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body models.UserResponse
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "Alice", body.Name)
}

func TestGetByID_NotFound(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	mockSvc.On("GetByID", mock.Anything, int32(99)).
		Return(models.UserResponse{}, service.ErrUserNotFound)

	req, _ := http.NewRequest("GET", "/api/users/99", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetByID_InvalidID(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	req, _ := http.NewRequest("GET", "/api/users/abc", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdate_Success(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	expected := models.UserResponse{ID: 1, Name: "Alice Updated", DOB: "1991-06-20", Age: 33}
	mockSvc.On("Update", mock.Anything, int32(1),
		models.UpdateUserRequest{Name: "Alice Updated", DOB: "1991-06-20"}).
		Return(expected, nil)

	req, _ := http.NewRequest("PUT", "/api/users/1",
		strings.NewReader(`{"name":"Alice Updated","dob":"1991-06-20"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body models.UserResponse
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "Alice Updated", body.Name)
}

func TestUpdate_NotFound(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	mockSvc.On("Update", mock.Anything, int32(99), mock.Anything).
		Return(models.UserResponse{}, service.ErrUserNotFound)

	req, _ := http.NewRequest("PUT", "/api/users/99",
		strings.NewReader(`{"name":"Alice","dob":"1990-05-15"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestUpdate_ValidationError(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	req, _ := http.NewRequest("PUT", "/api/users/1",
		strings.NewReader(`{"name":"","dob":"1990-05-15"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestDelete_Success(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	mockSvc.On("Delete", mock.Anything, int32(1)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/users/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestDelete_NotFound(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	mockSvc.On("Delete", mock.Anything, int32(99)).
		Return(service.ErrUserNotFound)

	req, _ := http.NewRequest("DELETE", "/api/users/99", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestDelete_InvalidID(t *testing.T) {
	mockSvc := new(mockUserService)
	app := setupTestApp(mockSvc)

	req, _ := http.NewRequest("DELETE", "/api/users/abc", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
