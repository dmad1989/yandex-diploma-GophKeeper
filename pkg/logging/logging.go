package logging

import (
	"fmt"

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
