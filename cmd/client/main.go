package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmad1989/gophKeeper/internal/config"
	"github.com/dmad1989/gophKeeper/pkg/logger"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
)

func main() {
	zlog, err := logger.NewLogger()
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

	fmt.Println(cfg.String())
	fmt.Println("client is working!")
}
