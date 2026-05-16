package middleware

/*
 * @Desc: 鉴权中间件
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"errors"
	"strings"
	"time"

	"github.com/gofurry/gofurry-user/apps/user/models"
	"github.com/gofurry/gofurry-user/common"
	"github.com/gofurry/gofurry-user/common/log"
	cs "github.com/gofurry/gofurry-user/common/service"
	"github.com/gofurry/gofurry-user/common/util"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleWare Fiber版本JWT鉴权中间件
func JWTMiddleWare() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authorization
		authorization := strings.TrimSpace(c.Get("Authorization"))
		if authorization == "" {
			// 返回未登录错误（Fiber使用JSON直接响应）
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    fiber.StatusUnauthorized,
				"message": "用户未登录.",
			})
		}

		// 从Redis获取有效token
		token := authorization
		cache, err := cs.GetString("jwt:" + token)
		if err != nil || strings.TrimSpace(cache) == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    fiber.StatusUnauthorized,
				"message": "登录信息已过期.",
			})
		}
		token = cache // 使用Redis中的有效token

		// 解析JWT token
		claims, pe := util.ParseToken(token)
		if pe != nil {
			log.Error(pe)
			// 根据错误类型返回对应信息
			switch {
			case errors.Is(pe, jwt.ErrTokenMalformed):
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code":    fiber.StatusUnauthorized,
					"message": "用户登录信息解析失败.",
				})
			case errors.Is(pe, jwt.ErrTokenSignatureInvalid):
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code":    fiber.StatusUnauthorized,
					"message": "用户登录信息失效.",
				})
			case errors.Is(pe, jwt.ErrTokenExpired), errors.Is(pe, jwt.ErrTokenNotValidYet):
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code":    fiber.StatusUnauthorized,
					"message": "用户登录信息已过期.",
				})
			default:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code":    fiber.StatusUnauthorized,
					"message": "用户未登录.",
				})
			}
		}

		// Redis续租
		if claims.ExpiresAt.Sub(time.Now()) < common.JWT_RELET_NUM*time.Hour {
			newTokenStr, err := util.NewToken(claims.UserId, claims.UserName)
			if err == nil {
				cs.SetExpire(authorization, newTokenStr, common.JWT_RELET_NUM*time.Hour)
			}
		}

		// 设置当前用户信息到上下文
		currentId, _ := util.String2Int64(claims.UserId)
		userInfo := models.CurrentUser{
			ID:   currentId,
			Name: claims.UserName,
		}
		c.Locals(common.COMMON_AUTH_CURRENT, userInfo)

		return c.Next()
	}
}
