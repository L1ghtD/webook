package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/ra1n6ow/webook/internal/repository"
	"github.com/ra1n6ow/webook/internal/repository/dao"
	"github.com/ra1n6ow/webook/internal/service"
	"github.com/ra1n6ow/webook/internal/web"
	"github.com/ra1n6ow/webook/internal/web/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	initDB()
	r := initWebServer()
	initUserHandler(r)
	r.Run(":8080")
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:123456@tcp(localhost:3306)/webook"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initUserHandler(server *gin.Engine) {
	db := initDB()
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	uh := web.NewUserHandler(us)
	uh.RegisterRoutes(server)
}

func initWebServer() *gin.Engine {
	r := gin.Default()

	// 跨域
	r.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 这里需要暴露出自定义头
		ExposeHeaders: []string{"X-Jwt-Token"},
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

	//useSession(r)
	useJWT(r)

	return r
}

func useJWT(r *gin.Engine) {
	login := middleware.LoginJWTMiddlewareBuilder{}
	r.Use(login.CheckLogin())
}

func useSession(r *gin.Engine) {
	login := &middleware.LoginMiddlewareBuilder{}
	// 存储数据的，也就是你 userId 存哪里
	// 直接存 cookie
	store := cookie.NewStore([]byte("secret"))
	//store, err := redis.NewStore(16, "tcp", "127.0.0.1:6379", "", []byte("aaabbbccc"))
	//if err != nil {
	//	panic(err)
	//}
	r.Use(sessions.Sessions("ssid", store), login.CheckLogin())
}
