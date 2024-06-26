package content

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
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

func (a *ContentApp) Save(ctx context.Context, c *model.Content) (err error) {
	if err = a.validateContent(c); err != nil {
		a.log.Errorw("ContentApp.Save: input object not valid",
			zap.Error(err))
		return fmt.Errorf("ContentApp.Save: validateContent %w", err)
	}

	c.ID, err = a.repo.SaveContent(ctx, *c)
	if err != nil {
		return fmt.Errorf("ContentApp.Save:  %w", err)
	}
	return nil
}

func (a *ContentApp) Update(ctx context.Context, c *model.Content) (err error) {
	if err = a.validateContent(c); err != nil {
		a.log.Errorw("ContentApp.Update: input object not valid",
			zap.Error(err))
		return fmt.Errorf("ContentApp.Update: validateContent: %w", err)
	}

	if err = a.repo.UpdateContent(ctx, c); err != nil {
		return fmt.Errorf("ContentApp.Update: %w", err)
	}
	return nil
}

func (a *ContentApp) Delete(ctx context.Context, id int32) error {
	if id == 0 {
		return errors.New("ContentApp.Delete: id is empty")
	}

	if err := a.repo.DeleteContent(ctx, id); err != nil {
		return fmt.Errorf("ContentApp.Delete: %w", err)
	}
	return nil
}

func (a *ContentApp) GetUserContent(ctx context.Context, typeID enum.ContentType) ([]*model.Content, error) {
	if typeID == enum.Nan {
		res, err := a.repo.GetAllUserContent(ctx)
		if err != nil {
			return nil, fmt.Errorf("ContentApp.GetUserContent: contenttype = nan: %w", err)
		}
		return res, nil
	}

	res, err := a.repo.GetUserContentByType(ctx, int32(typeID))
	if err != nil {
		return nil, fmt.Errorf("ContentApp.GetUserContent: contenttype != nan: %w", err)
	}
	return res, nil
}

func (a ContentApp) Get(ctx context.Context, id int32) (*model.Content, error) {
	if id == 0 {
		return nil, errors.New("ContentApp.Get: id  is iempty")
	}
	res, err := a.repo.GetUserContentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ContentApp.Get: %w", err)
	}
	return res, nil
}

func (a ContentApp) validateContent(c *model.Content) error {
	if len(c.Data) == 0 {
		return errors.New("content.Data in iempty")
	}
	if c.Meta == "" {
		return errors.New("content.Meta in iempty")
	}
	if c.Type == 0 {
		return errors.New("content.Type in iempty")
	}
	if c.UserID == 0 {
		return errors.New("content.UserID in iempty")
	}
	return nil
}
