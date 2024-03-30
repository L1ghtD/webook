package service

import (
	"context"
	"errors"
	"github.com/l1ghtd/webook/internal/domain"
	"github.com/l1ghtd/webook/internal/repository"
	"github.com/l1ghtd/webook/internal/repository/dao"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = dao.ErrDuplicateEmail
	ErrUserNotFound          = dao.ErrRecordNotFound
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo,
	}
}

func (us *UserService) SignUp(ctx context.Context, u domain.User) error {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(encrypted)
	return us.repo.Create(ctx, u)
}

func (us *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := us.repo.FindUserByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (us *UserService) Edit(ctx context.Context, u domain.User) error {
	return us.repo.Edit(ctx, u)
}

func (us *UserService) Profile(ctx context.Context, id int64) (dao.User, error) {
	u, err := us.repo.Query(ctx, id)
	return u, err
}
