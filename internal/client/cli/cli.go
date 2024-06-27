package cli

import (
	"context"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
)

type Auth interface {
	Register(ctx context.Context, username, password string) (*pb.TokenData, error)
	Login(ctx context.Context, username, password string) (*pb.TokenData, error)
}

type Content interface {
}

type cli struct {
	log     *zap.SugaredLogger
	auth    Auth
	content Content
}

func New(ctx context.Context, a Auth, c Content) *cli {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("CLI")
	return &cli{log: l, auth: a, content: c}
}

//TODO.....
