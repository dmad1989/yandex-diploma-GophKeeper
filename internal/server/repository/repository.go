package repository

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/dmad1989/gophKeeper/internal/server/repository/db"
	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var (
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

type Config interface {
	GetDBConn() string
}

type repo struct {
	logger  *zap.SugaredLogger
	queries *db.Queries
	dbConn  *sql.DB
}

func New(ctx context.Context, c Config) (*repo, error) {
	log := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("repository")

	pconf, err := pgxpool.ParseConfig(c.GetDBConn())
	if err != nil {
		return nil, fmt.Errorf("repository.new: ParseConfig: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pconf)
	if err != nil {
		return nil, fmt.Errorf("repository.new: pgxpool.NewWithConfig: %w", err)
	}
	dbConn := stdlib.OpenDBFromPool(pool)
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("goose.SetDialect: %w", err)
	}
	if err := goose.Up(dbConn, "migrations"); err != nil {
		return nil, fmt.Errorf("goose: create table: %w", err)
	}
	log.Debug("db connected")
	return &repo{
		log,
		db.New(pool),
		dbConn,
	}, nil
}

func (r repo) Close(ctx context.Context) (err error) {
	err = r.dbConn.Close()
	if err != nil {
		return fmt.Errorf("repository.Close: dbConn.Close: %w", err)
	}
	r.logger.Debug("db closed")
	return
}

func (r repo) CreateUser(ctx context.Context, u *model.User) (int32, error) {
	id, err := r.queries.CreateUser(ctx,
		db.CreateUserParams{
			Login:    u.Login,
			Password: u.HashPassword,
		})
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == pgerrcode.UniqueViolation {
			return 0, errs.ErrUserAlreadyExist
		}
		return 0, fmt.Errorf("repository.CreateUser: queries: %w", err)
	}
	return id, nil
}

func (r repo) GetUser(ctx context.Context, login string) (*model.User, error) {
	u, err := r.queries.GetUser(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.GetUser: queries: %w", err)
	}
	return &model.User{ID: u.ID, Login: u.Login, HashPassword: u.Password}, nil
}

func (r repo) SaveContent(ctx context.Context, c model.Content) (int32, error) {
	id, err := r.queries.SaveContent(ctx,
		db.SaveContentParams{
			UserID: c.UserID,
			Type:   int32(c.Type),
			Data:   c.Data,
			Meta:   pgtype.Text{String: c.Meta, Valid: true},
			Desc:   c.Desc,
		})

	if err != nil {
		return 0, fmt.Errorf("repository.SaveContent: queries: %w", err)
	}
	return id, nil
}

func (r repo) GetUserContentByID(ctx context.Context, id int32) (*model.Content, error) {
	userID, err := getUserIdFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.UpdateContent: getUserIdFromContext: %w", err)
	}
	dbc, err := r.queries.GetUserContentByID(ctx, db.GetUserContentByIDParams{ID: id, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("repository.UpdateContent: queries: %w", err)
	}
	return &model.Content{
		ID:     dbc.ID,
		UserID: dbc.UserID,
		Type:   enum.ContentType(dbc.Type),
		Data:   dbc.Data,
		Meta:   dbc.Meta.String,
		Desc:   dbc.Desc,
	}, nil
}

func (r repo) UpdateContent(ctx context.Context, c *model.Content) error {
	err := r.queries.UpdateContent(ctx, db.UpdateContentParams{
		ID:     c.ID,
		UserID: c.UserID,
		Type:   int32(c.Type),
		Data:   c.Data,
		Meta:   pgtype.Text{String: c.Meta, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("repository.UpdateContent: queries: %w", err)
	}
	return nil
}

func (r repo) GetUserContentByType(ctx context.Context, t int32) ([]*model.Content, error) {
	userID, err := getUserIdFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserContentByType: getUserIdFromContext: %w", err)
	}

	dbc, err := r.queries.GetUserContentByType(ctx, db.GetUserContentByTypeParams{UserID: userID, Type: t})
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserContentByType: queries: %w", err)
	}

	res := make([]*model.Content, len(dbc))

	for _, c := range dbc {
		res = append(res, &model.Content{
			ID:     c.ID,
			UserID: c.UserID,
			Type:   enum.ContentType(c.Type),
			Data:   c.Data,
			Meta:   c.Meta.String,
			Desc:   c.Desc,
		})
	}
	return res, nil
}

func (r repo) GetAllUserContent(ctx context.Context) ([]*model.Content, error) {
	userID, err := getUserIdFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.GetAllUserContent: getUserIdFromContext: %w", err)
	}
	dbc, err := r.queries.GetAllUserContent(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetAllUserContent: queries: %w", err)
	}

	res := make([]*model.Content, len(dbc))

	for _, c := range dbc {
		res = append(res, &model.Content{
			ID:     c.ID,
			UserID: c.UserID,
			Type:   enum.ContentType(c.Type),
			Data:   c.Data,
			Meta:   c.Meta.String,
			Desc:   c.Desc,
		})
	}
	return res, nil

}
func (r repo) DeleteContent(ctx context.Context, id int32) (err error) {
	err = r.queries.DeleteContent(ctx, id)
	if err != nil {
		err = fmt.Errorf("repository.DeleteContent: queries: %w", err)
	}
	return
}

func getUserIdFromContext(ctx context.Context) (int32, error) {
	userIDCtx := ctx.Value(consts.UserCtxKey)
	if userIDCtx == "" {
		return 0, errs.ErrNoCtxUser
	}
	userID, ok := userIDCtx.(int32)
	if !ok {
		return 0, errs.ErrNotIntCtxUser
	}
	return userID, nil
}
