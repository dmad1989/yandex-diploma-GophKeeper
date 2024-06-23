package user

import (
	"context"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"go.uber.org/zap"
)

type Repository interface {
}

type UserApp struct {
	log  *zap.SugaredLogger
	repo Repository
}

func NewApp(ctx context.Context, r Repository) *UserApp {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("UserApp")
	return &UserApp{log: l, repo: r}
}
