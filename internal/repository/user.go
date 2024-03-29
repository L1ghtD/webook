package repository

import (
	"context"
	"github.com/l1ghtd/webook/internal/domain"
	"github.com/l1ghtd/webook/internal/repository/dao"
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
	return ur.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}
