package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string
	Nickname string
	Birthday int64
	Intro    string

	// UTC
	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db,
	}
}

func (ud *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := ud.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		// 1062 表示重复代码
		if me.Number == duplicateErr {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (ud *UserDAO) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (ud *UserDAO) Update(ctx context.Context, u User) (User, error) {
	return u, nil
}
