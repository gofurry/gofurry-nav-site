package common

import (
	"net/http"
	"reflect"

	"github.com/gofiber/fiber/v3"
)

/*
 * @Desc: 响应 Code+Data
 * @author: 福狼
 * @version: v1.0.0
 */

// FiberResponse 响应结构体
type response struct {
	context fiber.Ctx
}

type ResultData struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

// NewFiberResponse 创建响应实例
func NewResponse(ctx fiber.Ctx) *response {
	return &response{
		context: ctx,
	}
}

func (r *response) Success() error {
	return r.context.Status(http.StatusOK).JSON(ResultData{
		Code: RETURN_SUCCESS,
		Data: "操作成功",
	})
}

// SuccessWithData 带数据的成功响应
func (r *response) SuccessWithData(data interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			// 发生恐慌时直接返回数据
			r.context.JSON(ResultData{
				Code: RETURN_SUCCESS,
				Data: data,
			})
		}
	}()

	// 检查数据是否有ID字段且值为0（与Gin版本逻辑一致）
	_, exist := reflect.TypeOf(data).FieldByName("ID")
	if exist {
		value := reflect.ValueOf(data).FieldByName("ID")
		id, ok := value.Interface().(int64)
		if id == 0 && ok {
			return r.context.JSON(ResultData{
				Code: RETURN_SUCCESS,
				Data: nil,
			})
		}
	}

	return r.context.JSON(ResultData{
		Code: RETURN_SUCCESS,
		Data: data,
	})
}

// 错误响应
func (r *response) Error(data interface{}) error {
	result := ResultData{
		Code: RETURN_FAILED,
		Data: data,
	}
	return r.context.JSON(result)
}

// 带状态码的错误响应
func (r *response) ErrorWithCode(data interface{}, code int) error {
	result := ResultData{
		Code: RETURN_FAILED,
		Data: data,
	}
	return r.context.Status(code).JSON(result)
}
