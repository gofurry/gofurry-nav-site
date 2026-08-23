package models

/*
 * @Desc: 鉴权
 * @author: 福狼
 * @version: v1.0.0
 */

import "github.com/golang-jwt/jwt/v5"

type GFClaims struct {
	jwt.RegisteredClaims
	UserName string `json:"userName"`
	UserId   string `json:"userId"`
}
