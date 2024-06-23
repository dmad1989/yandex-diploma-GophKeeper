package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmad1989/gophKeeper/internal/config"
	"github.com/dmad1989/gophKeeper/internal/server/app/content"
	"github.com/dmad1989/gophKeeper/internal/server/app/user"
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
	contectApp := content.NewApp(ctx, repo)

	zlog.Debug(contectApp)
	zlog.Debug(userApp)
	// TODO delete
	err = repo.Close(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("server is working!")
}
