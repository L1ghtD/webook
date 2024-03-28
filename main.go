package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/l1ghtd/webook/internal/web"
	"strings"
	"time"
)

func main() {
	r := initWebServer()
	initUserHandler(r)
	r.Run(":8080")
}

func initUserHandler(server *gin.Engine) {
	uh := web.NewUserHandler()
	uh.RegisterRoutes(server)
}

func initWebServer() *gin.Engine {
	r := gin.Default()

	// 跨域
	r.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		AllowHeaders: []string{"Content-Type"},
		//AllowHeaders: []string{"content-type"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				//if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	return r
}
