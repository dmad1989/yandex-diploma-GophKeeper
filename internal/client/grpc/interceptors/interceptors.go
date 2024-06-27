package interceptors

import (
	"context"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestTokenProvider struct {
	log         *zap.SugaredLogger
	tokenHolder *model.TokenHolder
}

func NewRequestTokenProvider(ctx context.Context, th *model.TokenHolder) *requestTokenProvider {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("RequestTokenProvider")
	return &requestTokenProvider{log: l, tokenHolder: th}
}

func (tp *requestTokenProvider) TokenInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(tp.ctxWithToken(ctx), method, req, reply, cc, opts...)
	}
}

func (tp *requestTokenProvider) TokenStreamInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(tp.ctxWithToken(ctx), desc, cc, method, opts...)
	}
}

func (tp *requestTokenProvider) ctxWithToken(ctx context.Context) context.Context {
	token := tp.tokenHolder.Get()
	if token != "" {
		return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"token": token}))
	}
	return ctx
}
