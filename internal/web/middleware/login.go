package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Gin 的 session 机制使用了 GOB 来将对象转化为字节切片[]byte，所以需要提前注册一下 Gob。
		gob.Register(time.Now())

		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			// 不需要登录校验
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			// 中断，不要往后执行，也就是不要执行后面的业务逻辑
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		// 刷新策略：每分钟刷新一次
		const updateTimeKey = "update_time"
		// 拿出上一次的刷新时间
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if !ok || val == nil || now.Sub(lastUpdateTime) > time.Second*60 {
			sess.Set(updateTimeKey, now)
			// 上面的 set 操作会把原有的 sess 所有 key 给覆盖掉，所以所有的 key 需要重新设置
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				fmt.Println(err)
			}

		}
	}
}
