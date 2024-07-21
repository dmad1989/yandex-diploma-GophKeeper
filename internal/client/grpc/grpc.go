package grpc

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/internal/client/grpc/interceptors"
	"github.com/dmad1989/gophKeeper/pkg/model/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Configer interface {
	GetServerAddress() string
}

func NewConnection(ctx context.Context, cfg Configer, th *client.TokenHolder) (*grpc.ClientConn, error) {
	tp := interceptors.NewRequestTokenProvider(ctx, th)
	creds, err := loadTLSCredentials()
	if err != nil {
		return nil, fmt.Errorf("grpc.NewConnection: %w", err)
	}
	conn, err := grpc.NewClient(
		cfg.GetServerAddress(),
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(tp.TokenInterceptor()),
		grpc.WithStreamInterceptor(tp.TokenStreamInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("grpc.NewConnection: NewClient: %w", err)
	}
	return conn, nil
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	creds, err := credentials.NewClientTLSFromFile("cert/server-cert.pem", "")
	if err != nil {
		return nil, fmt.Errorf("loadTLSCredentials: credentials.NewClientTLSFromFile: %w", err)
	}

	return creds, nil
}
