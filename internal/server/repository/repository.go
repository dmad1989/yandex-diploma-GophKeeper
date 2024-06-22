package repository

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/internal/server/repository/db"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type repo struct {
	logger  *zap.Logger
	queries *db.Queries
	dbConn  *pgx.Conn
}

func New(ctx context.Context, c string) (*repo, error) {
	log := ctx.Value(consts.LoggerCtxKey).(*zap.Logger).Named("repository")

	pconf, err := pgx.ParseConfig(c)
	if err != nil {
		return nil, fmt.Errorf("repository.new: ParseConfig: %w", err)
	}

	conn, err := pgx.ConnectConfig(ctx, pconf)
	if err != nil {
		return nil, fmt.Errorf("repository.new: ConnectConfig: %w", err)
	}

	return &repo{
		log,
		db.New(conn),
		conn,
	}, nil
}
