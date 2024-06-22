package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmad1989/gophKeeper/internal/config"
	"github.com/dmad1989/gophKeeper/tools/logging"
	"github.com/dmad1989/gophKeeper/tools/model/consts"
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

	fmt.Println(cfg.String())
	fmt.Println("client is working!")
}
