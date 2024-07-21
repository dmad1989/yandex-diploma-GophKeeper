package test

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"go.uber.org/zap"
)

type Context context.Context

type builder struct {
	logger *zap.SugaredLogger
	userId int32
}

func (b *builder) initLogger() (*builder, error) {
	zl, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("loggerInit: %w", err)
	}
	b.logger = zl.Sugar()
	return b, nil
}

func (b *builder) setUserId(id int32) *builder {
	b.userId = id
	return b
}

func (b *builder) build() (ctx Context) {
	ctx = context.Background()
	if b.logger != nil {
		ctx = context.WithValue(ctx, consts.LoggerCtxKey, b.logger)
	}
	if b.userId != 0 {
		ctx = context.WithValue(ctx, consts.UserCtxKey, b.userId)
	}
	return
}

func newBuilder() *builder {
	return &builder{}
}
func NewContextFull() (Context, error) {
	b, err := newBuilder().setUserId(10).initLogger()
	if err != nil {
		return nil, fmt.Errorf("testContext.NewFullContext: builder.initLogger: %w", err)
	}
	return b.build(), nil
}

func NewContextNoLogger() Context {
	return newBuilder().setUserId(10).build()
}

func NewContextNoUserId() (Context, error) {
	b, err := newBuilder().initLogger()
	if err != nil {
		return nil, fmt.Errorf("testContext.NewFullContext: builder.initLogger: %w", err)
	}
	return b.build(), nil
}

func NewContextEmpty() Context {
	return newBuilder().build()
}
