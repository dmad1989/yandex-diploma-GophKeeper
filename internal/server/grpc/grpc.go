package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/dmad1989/gophKeeper/internal/server/grpc/interceptors"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Configer interface {
	GetServerAddress() string
}

type TokenProvider interface {
	TokenInterceptor() grpc.UnaryServerInterceptor
}

type Servers struct {
	grpc     *grpc.Server
	log      *zap.SugaredLogger
	auth     pb.AuthServer
	contents pb.ContentsServer
	cfg      Configer
}

type ServerBuilder struct {
	ctx      context.Context
	auth     pb.AuthServer
	contents pb.ContentsServer
	cfg      Configer
	app      UserApp
}

func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{}
}

func (sb *ServerBuilder) Context(c context.Context) *ServerBuilder {
	sb.ctx = c
	return sb
}

func (sb *ServerBuilder) AuthServer(a pb.AuthServer) *ServerBuilder {
	sb.auth = a
	return sb
}

func (sb *ServerBuilder) ContentsServer(c pb.ContentsServer) *ServerBuilder {
	sb.contents = c
	return sb
}
func (sb *ServerBuilder) Config(c Configer) *ServerBuilder {
	sb.cfg = c
	return sb
}
func (sb *ServerBuilder) UserApp(u UserApp) *ServerBuilder {
	sb.app = u
	return sb
}

func (sb ServerBuilder) validate() error {
	if sb.app == nil {
		return errors.New("validate: empty UserApp")
	}
	if sb.auth == nil {
		return errors.New("validate: empty AuthServ")
	}
	if sb.cfg == nil {
		return errors.New("validate: empty Config")
	}
	if sb.contents == nil {
		return errors.New("validate: empty ContentsServ")
	}
	if sb.ctx == nil {
		return errors.New("validate: empty Context")
	}
	return nil
}

func (sb *ServerBuilder) Build() (*Servers, error) {
	if err := sb.validate(); err != nil {
		return nil, fmt.Errorf("ServerBuilder.build: %w", err)
	}
	l := sb.ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("grpc")
	tp := interceptors.NewTokenProvider(sb.ctx, sb.app, "")
	s := grpc.NewServer(
		grpc.UnaryInterceptor(tp.TokenInterceptor()),
		grpc.StreamInterceptor(tp.TokenStreamInterceptor()),
	)
	return &Servers{
		grpc:     s,
		log:      l,
		auth:     sb.auth,
		contents: sb.contents,
		cfg:      sb.cfg}, nil
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
