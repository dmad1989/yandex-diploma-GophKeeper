package main

import (
	"context"
	"log"

	"github.com/dmad1989/gophKeeper/internal/client/app/auth"
	"github.com/dmad1989/gophKeeper/internal/client/app/content"
	"github.com/dmad1989/gophKeeper/internal/client/app/crypto"
	"github.com/dmad1989/gophKeeper/internal/client/cli"
	"github.com/dmad1989/gophKeeper/internal/client/grpc"
	"github.com/dmad1989/gophKeeper/internal/config"
	"github.com/dmad1989/gophKeeper/pkg/logging"
	"github.com/dmad1989/gophKeeper/pkg/model/client"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
)

func main() {
	zlog, err := logging.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	zlog = zlog.Named("client")
	ctx := context.WithValue(context.Background(), consts.LoggerCtxKey, zlog)
	defer zlog.Sync()

	cfg, err := config.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tokenHolder := &client.TokenHolder{}

	conn, err := grpc.NewConnection(ctx, cfg, tokenHolder)
	if err != nil {
		log.Fatal(err)
	}
	authApp := auth.New(ctx, conn, tokenHolder)
	cryptoApp := crypto.New(ctx, cfg)
	contentApp := content.New(ctx, conn, cryptoApp)
	//TODO
	cli.New(ctx, authApp, contentApp)
	<-ctx.Done()
}
