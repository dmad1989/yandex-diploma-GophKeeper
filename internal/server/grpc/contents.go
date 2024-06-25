package grpc

import (
	"context"

	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
)

type Contents struct {
}

func NewContentsServer(ctx context.Context, c any) pb.ContentsServer {
	return pb.UnimplementedContentsServer{}
}
