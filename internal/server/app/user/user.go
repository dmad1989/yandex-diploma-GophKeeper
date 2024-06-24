package user

import (
	"context"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"go.uber.org/zap"
)

type Repository interface {
	CreateUser(ctx context.Context, u model.User) (int32, error)
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

