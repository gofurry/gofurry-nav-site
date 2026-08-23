package common

import "net/http"

type Error interface {
	error
	GetErrorCode() int
	GetMsg() string
	GetHTTPStatus() int
}

type appError struct {
	errorCode  int
	httpStatus int
	msg        string
}

func NewError(errorCode, httpStatus int, msg string) *appError {
	return &appError{
		errorCode:  errorCode,
		httpStatus: httpStatus,
		msg:        msg,
	}
}

func NewServiceError(msg string) *appError {
	return NewError(RETURN_FAILED, http.StatusInternalServerError, msg)
}

func NewDaoError(msg string) *appError {
	return NewError(RETURN_FAILED, http.StatusInternalServerError, msg)
}

func NewValidationError(msg string) *appError {
	return NewError(RETURN_FAILED, http.StatusBadRequest, msg)
}

func (ae appError) Error() string {
	return ae.msg
}

func (ae appError) GetErrorCode() int {
	return ae.errorCode
}

func (ae appError) GetMsg() string {
	return ae.msg
}

func (ae appError) GetHTTPStatus() int {
	return ae.httpStatus
}
