package interceptors

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type UserApp interface {
	ExtractIDFromToken(t string) (int32, error)
}

type TokenProvider struct {
	log             *zap.SugaredLogger
	app             UserApp
	nonSecureMethod map[string]struct{}
}

func NewTokenProvider(ctx context.Context, a UserApp, methods ...string) *TokenProvider {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("TokenProvider")
	m := make(map[string]struct{}, len(methods))

	for _, method := range methods {
		m[method] = struct{}{}
	}

	return &TokenProvider{log: l, app: a, nonSecureMethod: m}
}

func (tp *TokenProvider) isSecureMethod(method string) bool {
	if _, ok := tp.nonSecureMethod[method]; ok {
		return true
	}
	return false
}

func (tp *TokenProvider) TokenInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if !tp.isSecureMethod(info.FullMethod) {
			userId, err := tp.extractID(ctx)
			if err != nil {
				return nil, errs.TokenError{Err: err}
			}
			ctxWithUserId := context.WithValue(ctx, consts.UserCtxKey, userId)
			return handler(ctxWithUserId, req)
		}
		return handler(ctx, req)
	}
}

func (tp *TokenProvider) extractID(ctx context.Context) (int32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("TokenProvider.extractID: metadata.FromIncomingContext: %w", errs.ErrReadMD)
	}
	var tokenStr string
	if values := md.Get("token"); len(values) == 0 {
		return 0, fmt.Errorf("TokenProvider.extractID: md.Get: %w", errs.ErrTokenNotFound)
	} else {
		tokenStr = values[0]
	}
	id, err := tp.app.ExtractIDFromToken(tokenStr)
	if err != nil {
		return 0, fmt.Errorf("TokenProvider.extractID: %w", err)
	}
	return id, nil
}
