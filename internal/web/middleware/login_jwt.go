package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ra1n6ow/webook/internal/web"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			// 不需要登录校验
			return
		}

		// 根据约定，token 在 Authorization 头部
		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			// 没登录，没有 token, Authorization 这个头部都没有
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(authCode, " ")
		if len(segs) != 2 {
			// 没登录，Authorization 中的内容是乱传的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Bearer XXXX
		tokenStr := segs[1]

		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			// 这是一个计算JWTKey 的函数，因为已经提前定义了，所以直接返回
			return web.JWTKey, nil
		})
		if err != nil {
			// token 不对，token 是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			// token 解析出来了，但是 token 可能是非法的，或者过期了的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			// 和上一次的User-Agent不一样
			// 能够进来这个分支的，大概率是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime := uc.ExpiresAt
		// 不判定都可以
		//if expireTime.Before(time.Now()) {
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		// 剩余过期时间 < 50s 就要刷新
		if expireTime.Sub(time.Now()) < time.Second*50 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 5))
			// 重新签名
			tokenStr, err = token.SignedString(web.JWTKey)
			ctx.Header("x-jwt-token", tokenStr)
			if err != nil {
				// 这边不要中断，因为仅仅是过期时间没有刷新，但是用户是登录了的
				log.Println(err)
			}
		}
		ctx.Set("user", uc)
	}
}
