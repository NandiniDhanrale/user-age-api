package handler

import (
	"strconv"

	"github.com/NandiniDhanrale/user-age-api/internal/models"
	"github.com/NandiniDhanrale/user-age-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	svc service.UserService
	vld *validator.Validate
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
		vld: validator.New(),
	}
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if err := h.vld.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	user, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	user, err := h.svc.GetByID(c.Context(), int32(id))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	users, err := h.svc.List(c.Context())
	if err != nil {
		return err
	}

	if users == nil {
		users = []models.UserResponse{}
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if err := h.vld.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	user, err := h.svc.Update(c.Context(), int32(id), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	if err := h.svc.Delete(c.Context(), int32(id)); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
