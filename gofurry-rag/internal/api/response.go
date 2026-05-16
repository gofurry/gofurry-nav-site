package api

import (
	"errors"
	"net/http"

	"github.com/gofurry/gofurry-rag/internal/auth"
	"github.com/gofurry/gofurry-rag/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type statusCoder interface {
	HTTPStatus() int
}

func ok(c fiber.Ctx, data any) error {
	return c.Status(http.StatusOK).JSON(Result{Code: 1, Message: "success", Data: data})
}

func ErrorHandler(c fiber.Ctx, err error) error {
	return fail(c, err)
}

func ErrorWithCode(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Result{Code: 0, Message: message, Data: fiber.Map{}})
}

func fail(c fiber.Ctx, err error) error {
	status, message := errorStatus(err)
	switch {
	case errors.Is(err, service.ErrValidation):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, auth.ErrInvalidPassword), errors.Is(err, auth.ErrNotLoggedIn), errors.Is(err, auth.ErrInvalidSession):
		status = http.StatusUnauthorized
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

func errorStatus(err error) (int, string) {
	status := http.StatusInternalServerError
	message := "internal server error"
	var sc statusCoder
	switch {
	case errors.As(err, &sc):
		status = sc.HTTPStatus()
		message = err.Error()
	case err != nil:
		message = err.Error()
	}
	return status, message
}
