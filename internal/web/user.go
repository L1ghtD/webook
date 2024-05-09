package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ra1n6ow/webook/internal/domain"
	"github.com/ra1n6ow/webook/internal/service"
	"net/http"
	"strings"
	"time"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 用 ` 看起来比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	nicknameRegexPattern = `^.{2,5}$`
	introRegexPattern    = `^.{5,10}$`
	birthdayRegexPattern = `^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	nicknameRexExp *regexp.Regexp
	introRexExp    *regexp.Regexp
	birthdayRexExp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		nicknameRexExp: regexp.MustCompile(nicknameRegexPattern, regexp.None),
		introRexExp:    regexp.MustCompile(introRegexPattern, regexp.None),
		birthdayRexExp: regexp.MustCompile(birthdayRegexPattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", h.SignUp)
	// POST /users/login
	//ug.POST("/login", h.Login)
	ug.POST("/login", h.LoginJWT)
	// POST /users/edit
	ug.POST("/edit", h.Edit)
	// GET /users/profile
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = h.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十五分钟
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		// Header + Payload
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
			Uid:       u.Id,
			UserAgent: ctx.GetHeader("User-Agent"),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
		})
		// 使用 JWTKey 签名
		// HS256,HS384,HS512 签名可以直接传入 []byte, 其它签名方法需要看文档
		tokenStr, err := token.SignedString(JWTKey)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.Header("x-jwt-token", tokenStr)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Birthday string `json:"birthday"`
		Nickname string `json:"nickname"`
		Intro    string `json:"intro"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//fmt.Println(ctx.Param("id"))

	session := sessions.Default(ctx)
	var userId int64
	userId = session.Get("userId").(int64)

	isBirthday, err := h.birthdayRexExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isBirthday {
		ctx.String(http.StatusOK, "非法生日格式, 比如 1987-03-04")
		return
	}

	isNickname, err := h.nicknameRexExp.MatchString(req.Nickname)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isNickname {
		ctx.String(http.StatusOK, "长度不符合要求，2-5位")
		return
	}

	isIntro, err := h.introRexExp.MatchString(req.Intro)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isIntro {
		ctx.String(http.StatusOK, "长度不符合要求，5-10位")
		return
	}

	user := domain.User{
		Id:       userId,
		Birthday: parseTime(req.Birthday),
		Nickname: req.Nickname,
		Intro:    req.Intro,
	}
	err = h.svc.Edit(ctx, user)
	switch err {
	case nil:
		ctx.String(http.StatusOK, "修改成功")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) Profile(ctx *gin.Context) {
	type ProfileRsp struct {
		Id       int64  `json:"id"`
		Email    string `json:"email"`
		Birthday string `json:"birthday"`
		Nickname string `json:"nickname"`
		Intro    string `json:"intro"`
	}

	// session
	//session := sessions.Default(ctx)
	//var userId int64
	//userId = session.Get("userId").(int64)

	// JWT
	authCode := ctx.GetHeader("Authorization")
	segs := strings.Split(authCode, " ")
	tokenStr := segs[1]

	var uc UserClaims
	_, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
		// 这是一个计算JWTKey 的函数，因为已经提前定义了，所以直接返回
		return JWTKey, nil
	})
	userId := uc.Uid

	daoUser, err := h.svc.Profile(ctx, userId)
	switch err {
	case nil:
		profileRsp := ProfileRsp{
			Id:       daoUser.Id,
			Email:    daoUser.Email,
			Birthday: transUnixToStr(daoUser.Birthday),
			Nickname: daoUser.Nickname,
			Intro:    daoUser.Intro,
		}
		ctx.JSON(http.StatusOK, profileRsp)
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func parseTime(birthday string) time.Time {
	layout := "2006-01-02"
	t, _ := time.Parse(layout, birthday)
	return t
}

func transUnixToStr(unixTime int64) string {
	t := time.UnixMilli(unixTime)
	formattedTime := t.Format("2006-01-02")
	return formattedTime
}

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type UserClaims struct {
	// 该结构体实现了 Claims 接口
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
