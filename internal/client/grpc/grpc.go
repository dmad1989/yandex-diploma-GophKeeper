package grpc

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/internal/client/grpc/interceptors"
	"github.com/dmad1989/gophKeeper/pkg/model"
	"google.golang.org/grpc"
)

type Configer interface {
	GetServerAddress() string
}

func NewConnection(ctx context.Context, cfg Configer, th *model.TokenHolder) (*grpc.ClientConn, error) {
	tp := interceptors.NewRequestTokenProvider(ctx, th)
	conn, err := grpc.NewClient(
		cfg.GetServerAddress(),
		grpc.WithUnaryInterceptor(tp.TokenInterceptor()),
		grpc.WithStreamInterceptor(tp.TokenStreamInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("grpc.NewConnection: NewClient: %w", err)
	}
	return conn, nil
}
