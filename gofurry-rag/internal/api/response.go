package api

import (
	"errors"
	"net/http"

	"github.com/GoFurry/gofurry-rag/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func ok(c fiber.Ctx, data any) error {
	return c.Status(http.StatusOK).JSON(Result{Code: 1, Message: "success", Data: data})
}

func fail(c fiber.Ctx, err error) error {
	status := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, service.ErrValidation):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, pgx.ErrNoRows):
		status = http.StatusNotFound
		message = "resource not found"
	case errors.Is(err, fiber.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case err != nil:
		message = err.Error()
	}
	return c.Status(status).JSON(Result{Code: 0, Message: message, Data: fiber.Map{}})
}
