package common

/*
 * @Desc: 异常定义
 * @author: 福狼
 * @version: v1.0.0
 */

type GFError interface {
	GetErrorCode() int
	GetMsg() string
}

func NewServiceError(msg string) *serviceError {
	return &serviceError{
		errorCode: RETURN_FAILED,
		msg:       msg,
	}
}

type serviceError struct {
	errorCode int
	msg       string
}

func (se serviceError) GetErrorCode() int { return se.errorCode }

func (se serviceError) GetMsg() string { return se.msg }

func NewDaoError(msg string) *daoError {
	return &daoError{
		errorCode: RETURN_FAILED,
		msg:       msg,
	}
}

type daoError struct {
	errorCode int
	msg       string
}

func (de daoError) GetErrorCode() int { return de.errorCode }

func (de daoError) GetMsg() string { return de.msg }
