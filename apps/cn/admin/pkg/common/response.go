package common

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

type response struct {
	context fiber.Ctx
}

type ResultData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewResponse(ctx fiber.Ctx) *response {
	return &response{context: ctx}
}

func (r *response) Success() error {
	return r.write(http.StatusOK, RETURN_SUCCESS, "success", nil)
}

func (r *response) SuccessWithData(data any) error {
	return r.write(http.StatusOK, RETURN_SUCCESS, "success", data)
}

func (r *response) Error(data any) error {
	appErr := normalizeError(data)
	return r.write(appErr.GetHTTPStatus(), appErr.GetErrorCode(), appErr.GetMsg(), nil)
}

func (r *response) ErrorWithCode(data any, status int) error {
	appErr := normalizeError(data)
	return r.write(status, appErr.GetErrorCode(), appErr.GetMsg(), nil)
}

func (r *response) write(status, code int, message string, data any) error {
	return r.context.Status(status).JSON(ResultData{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func normalizeError(data any) Error {
	switch value := data.(type) {
	case nil:
		return NewServiceError("request failed")
	case Error:
		return value
	case error:
		return NewServiceError(value.Error())
	case string:
		return NewServiceError(value)
	default:
		return NewServiceError("request failed")
	}
}
