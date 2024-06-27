package auth

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type authApp struct {
	log         *zap.SugaredLogger
	client      pb.AuthClient
	tokenHolder *model.TokenHolder
}

func New(ctx context.Context, conn *grpc.ClientConn, t *model.TokenHolder) *authApp {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("Auth")
	return &authApp{log: l, client: pb.NewAuthClient(conn), tokenHolder: t}
}

func (a *authApp) Register(ctx context.Context, username, password string) (*pb.TokenData, error) {
	tokenData, err := a.client.Register(ctx, &pb.AuthData{Username: username, Password: password})

	if err != nil {
		return nil, fmt.Errorf("AuthApp.Register: client.Register: %w ", err)
	}

	a.tokenHolder.Set(tokenData.Token)
	return tokenData, nil
}

func (a *authApp) Login(ctx context.Context, username, password string) (*pb.TokenData, error) {
	tokenData, err := a.client.Login(ctx, &pb.AuthData{Username: username, Password: password})

	if err != nil {
		return nil, fmt.Errorf("AuthApp.Register: client.Login: %w ", err)
	}

	a.tokenHolder.Set(tokenData.Token)
	return tokenData, nil
}
