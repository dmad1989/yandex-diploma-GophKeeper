package main

import (
	"context"
	"fmt"
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
	zlog, err := logging.NewLogger()
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

	userApp := user.NewApp(ctx, repo)
	contentApp := content.NewApp(ctx, repo)

	zlog.Debug(contentApp)
	zlog.Debug(userApp)
	authSrv := grpc.NewAuthServer(ctx)
	contentsSrv := grpc.NewContentsServer(ctx)

	s := grpc.NewServer(ctx, authSrv, contentsSrv, cfg)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	s.Run(ctx)
	fmt.Println("server is working!")
	<-ctx.Done()
	s.Stop()
}
