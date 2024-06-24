package grpc

import (
	"context"
	"net"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Configer interface {
	GetServerAddress() string
}

type Servers struct {
	grpc     *grpc.Server
	log      *zap.SugaredLogger
	auth     pb.AuthServer
	contents pb.ContentsServer
	cfg      Configer
}

func NewServer(ctx context.Context, a pb.AuthServer, cont pb.ContentsServer, cfg Configer) *Servers {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("grpc")
	return &Servers{
		grpc:     grpc.NewServer(),
		log:      l,
		auth:     a,
		contents: cont,
		cfg:      cfg}
}

func (s *Servers) Run(ctx context.Context) {
	pb.RegisterAuthServer(s.grpc, s.auth)
	pb.RegisterContentsServer(s.grpc, s.contents)

	go func() {
		srv, err := net.Listen("tcp", s.cfg.GetServerAddress())
		if err != nil {
			s.log.Errorf("grpc.Run: net.Listen: %w", err)
		}
		s.log.Info("gRPC server started")
		err = s.grpc.Serve(srv)
		if err != nil {
			s.log.Errorf("grpc.Run: grpc.Serve: %w", err)
		}
	}()

}

func (s *Servers) Stop() {
	s.grpc.GracefulStop()
}

//TODO graceful shutdown
// https://medium.com/@pthtantai97/mastering-grpc-server-with-graceful-shutdown-within-golangs-hexagonal-architecture-0bba657b8622
