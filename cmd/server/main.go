package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmad1989/gophKeeper/pkg/logger"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
)

func main() {
	zlog, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	zlog = zlog.Named("server")
	ctx := context.WithValue(context.Background(), consts.LoggerCtxKey, zlog)
	defer zlog.Sync()

	fmt.Println("server is working!")
}
