package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmad1989/gophKeeper/internal/config"
	"github.com/dmad1989/gophKeeper/internal/server/repository"
	"github.com/dmad1989/gophKeeper/tools/logging"
	"github.com/dmad1989/gophKeeper/tools/model/consts"
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
	// TODO delete
	err = repo.Close(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("server is working!")
}
