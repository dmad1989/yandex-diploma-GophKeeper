package content

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/logging"
	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	"go.uber.org/zap"
)

//go:generate mockgen -source=content.go -destination=./mock_content.go -package=content

var (
	ErrEmptyID     = errors.New("content.ID is iempty")
	ErrEmptyData   = errors.New("content.Data is iempty")
	ErrEmptyMeta   = errors.New("content.Meta is iempty")
	ErrEmptyType   = errors.New("content.Type is iempty")
	ErrEmptyUserID = errors.New("content.UserID is iempty")
	ErrTypeToDesc  = errors.New("content.Type is not associated with Desc")
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

func NewApp(ctx context.Context, r Repository) (*ContentApp, error) {
	l, err := logging.LoggerFromContext(ctx, "ContentApp")
	if err != nil {
		return nil, fmt.Errorf("auth.NewAuthServer: LoggerFromContext: %w", err)
	}

	return &ContentApp{log: l, repo: r}, nil
}

func (a *ContentApp) Save(ctx context.Context, c *model.Content) (err error) {
	if err = a.validateContent(c); err != nil {
		a.log.Errorw("ContentApp.Save: input object not valid",
			zap.Error(err))
		return fmt.Errorf("ContentApp.Save: validateContent %w", err)
	}

	desc, ok := enum.TypeToDesc[c.Type]
	if !ok {
		return ErrTypeToDesc
	}

	c.Desc = []byte(desc)

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

	if c.ID == 0 {
		a.log.Errorw("ContentApp.Update: input object not valid",
			zap.Error(ErrEmptyID))
		return fmt.Errorf("ContentApp.Update: %w", ErrEmptyID)
	}

	if err = a.repo.UpdateContent(ctx, c); err != nil {
		return fmt.Errorf("ContentApp.Update: %w", err)
	}
	return nil
}

func (a *ContentApp) Delete(ctx context.Context, id int32) error {
	if id == 0 {
		return fmt.Errorf("ContentApp.Delete: %w", ErrEmptyID)
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
		return nil, fmt.Errorf("ContentApp.Get: %w", ErrEmptyID)
	}
	res, err := a.repo.GetUserContentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ContentApp.Get: %w", err)
	}
	return res, nil
}

func (a ContentApp) validateContent(c *model.Content) error {
	if len(c.Data) == 0 {
		return ErrEmptyData
	}
	if c.Meta == "" {
		return ErrEmptyMeta
	}
	if c.Type == 0 {
		return ErrEmptyType
	}
	if c.UserID == 0 {
		return ErrEmptyUserID
	}
	return nil
}
