package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmad1989/gophKeeper/internal/config"
	"github.com/dmad1989/gophKeeper/internal/server/app/content"
	"github.com/dmad1989/gophKeeper/internal/server/app/user"
	"github.com/dmad1989/gophKeeper/internal/server/grpc"
	"github.com/dmad1989/gophKeeper/internal/server/repository"
	"github.com/dmad1989/gophKeeper/pkg/logging"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
)

func main() {
	zlog, err := logging.NewLogger("")
	if err != nil {
		log.Fatal(err)
	}
	zlog = zlog.Named("server")
	ctx := context.WithValue(context.Background(), consts.LoggerCtxKey, zlog)
	defer zlog.Sync()

	cfg, err := config.NewServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	repo, err := repository.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close(ctx)
	userApp, err := user.NewApp(ctx, repo)
	if err != nil {
		log.Fatal(err)
	}
	contentApp, err := content.NewApp(ctx, repo)
	if err != nil {
		log.Fatal(err)
	}

	authSrv, err := grpc.NewAuthServer(ctx, userApp)
	if err != nil {
		log.Fatal(err)
	}
	contentsSrv, err := grpc.NewContentsServer(ctx, contentApp)
	if err != nil {
		log.Fatal(err)
	}

	s, err := grpc.NewServerBuilder().
		Context(ctx).
		AuthServer(authSrv).
		ContentsServer(contentsSrv).
		Config(cfg).
		UserApp(userApp).Build()

	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	s.Run(ctx)
	zlog.Info("server is working!")
	<-ctx.Done()
	s.Stop()
}
