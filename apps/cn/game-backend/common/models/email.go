package models

/*
 * @Desc: 邮箱
 * @author: 福狼
 * @version: v1.0.0
 */

type EmailCodeModel struct {
	Email string `form:"email" json:"email" validate:"required,email,min=0,max=255"`
	Code  string `form:"code" json:"code" validate:"required"`
}

type EmailCode struct {
	EmailCodeModel
}
