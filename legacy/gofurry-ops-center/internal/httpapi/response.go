package httpapi

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func ok(c fiber.Ctx, data any) error {
	return c.JSON(Result{Code: 1, Message: "success", Data: data})
}

func accepted(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusAccepted).JSON(Result{Code: 1, Message: "accepted", Data: data})
}

func fail(c fiber.Ctx, status int, message string) error {
	if message == "" {
		message = "request failed"
	}
	return c.Status(status).JSON(Result{Code: 0, Message: message})
}

func ErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		message = fiberErr.Message
	}
	return fail(c, code, message)
}
