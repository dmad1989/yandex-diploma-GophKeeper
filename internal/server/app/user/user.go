package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, u *model.User) (int32, error)
	GetUser(ctx context.Context, login string) (*model.User, error)
}

type UserApp struct {
	log  *zap.SugaredLogger
	repo Repository
}

func NewApp(ctx context.Context, r Repository) *UserApp {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("UserApp")
	return &UserApp{log: l, repo: r}
}

func (a *UserApp) Register(ctx context.Context, user *model.User) error {
	var err error
	user.HashPassword, err = bcrypt.GenerateFromPassword([]byte(user.Password), 15)
	if err != nil {
		return fmt.Errorf("User.Register: crypt.GenerateFromPassword: %w", err)

	}
	user.ID, err = a.repo.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("User.Register: %w", err)
	}
	return nil
}

func (s *UserApp) GetByLogin(ctx context.Context, login string) (user *model.User, err error) {
	user, err = s.repo.GetUser(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("User.GetByLogin: %w", err)
	}
	return
}

func (us *UserApp) ValidatePassword(cxt context.Context, user *model.User, password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(user.HashPassword, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("User.ValidatePassword: bcrypt.CompareHashAndPassword  %w", err)
	}
	return true, nil
}
