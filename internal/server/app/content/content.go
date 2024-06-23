package content

import (
	"context"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"go.uber.org/zap"
)

type Repository interface {
	SaveContent(ctx context.Context, c model.Content) (int32, error)
	GetUserContentByID(ctx context.Context, id int32) (*model.Content, error)
	GetUserContentByType(ctx context.Context, t int32) ([]*model.Content, error)
	UpdateContent(ctx context.Context, c *model.Content) error
	GetAllUserContent(ctx context.Context) ([]*model.Content, error)
	DeleteContent(ctx context.Context, id int32) (err error)
}

type ContentApp struct {
	log  *zap.SugaredLogger
	repo Repository
}

func NewApp(ctx context.Context, r Repository) *ContentApp {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("ContentApp")
	return &ContentApp{log: l, repo: r}
}
