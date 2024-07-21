package logging

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"go.uber.org/zap"
)

func NewLogger(logFile string) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	if logFile != "" {
		cfg.OutputPaths = []string{logFile}
	}
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("logger.NewLogger: cfg.Build: %w", err)
	}
	return logger.Sugar(), nil
}

func LoggerFromContext(ctx context.Context, name string) (*zap.SugaredLogger, error) {
	if ctx == nil {
		return nil, errs.ErrNoCtx
	}
	l := ctx.Value(consts.LoggerCtxKey)
	if l == nil {
		return nil, errs.ErrNoCtxLogger
	}
	return l.(*zap.SugaredLogger).Named(name), nil
}
