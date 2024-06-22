package repository

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/internal/server/repository/db"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Config interface {
	GetDBConn() string
}

type repo struct {
	logger  *zap.SugaredLogger
	queries *db.Queries
	dbConn  *pgx.Conn
}

func New(ctx context.Context, c Config) (*repo, error) {
	log := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("repository")

	pconf, err := pgx.ParseConfig(c.GetDBConn())
	if err != nil {
		return nil, fmt.Errorf("repository.new: ParseConfig: %w", err)
	}

	conn, err := pgx.ConnectConfig(ctx, pconf)
	if err != nil {
		return nil, fmt.Errorf("repository.new: ConnectConfig: %w", err)
	}
	log.Debug("db connected")
	return &repo{
		log,
		db.New(conn),
		conn,
	}, nil
}

func (r repo) Close(ctx context.Context) (err error) {
	err = r.dbConn.Close(ctx)
	if err != nil {
		return fmt.Errorf("repository.Close: dbConn.Close: %w", err)
	}
	r.logger.Debug("db closed")
	return
}
