package repository

import (
	"context"
	"github.com/ra1n6ow/webook/internal/domain"
	"github.com/ra1n6ow/webook/internal/repository/dao"
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao,
	}
}

func (ur *UserRepository) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (ur *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := ur.dao.FindUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return ur.toDomainUser(u), nil
}

func (ur *UserRepository) toDomainUser(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}

func (ur *UserRepository) Edit(ctx context.Context, u domain.User) error {
	return ur.dao.Update(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday.UnixMilli(),
		Intro:    u.Intro,
	})
}

func (ur *UserRepository) Query(ctx context.Context, id int64) (dao.User, error) {
	u, err := ur.dao.Query(ctx, dao.User{
		Id: id,
	})
	return u, err
}
