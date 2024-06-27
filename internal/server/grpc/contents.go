package grpc

import (
	"context"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ContentApp interface {
	Save(ctx context.Context, c *model.Content) (err error)
	Update(ctx context.Context, c *model.Content) (err error)
	Delete(ctx context.Context, id int32) error
	GetUserContent(ctx context.Context, typeID enum.ContentType) ([]*model.Content, error)
	Get(ctx context.Context, id int32) (*model.Content, error)
}

type contentsServer struct {
	log *zap.SugaredLogger
	pb.UnimplementedContentsServer
	app ContentApp
}

func NewContentsServer(ctx context.Context, c ContentApp) pb.ContentsServer {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("ContentsServer")
	return &contentsServer{log: l, app: c}
}

func (c *contentsServer) Save(ctx context.Context, content *pb.Content) (*pb.ContentId, error) {
	userID, err := getUserIdFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("contentServ.Save: getUserIdFromContext: %w", err)
	}
	mContent := &model.Content{
		UserID: userID,
		Type:   enum.ContentType(content.Type),
		Data:   content.Data,
		Meta:   content.Meta,
	}
	if err := c.app.Save(ctx, mContent); err != nil {
		return nil, status.Errorf(codes.Internal, "contentServ.Save: %s", err.Error())
	}
	return &pb.ContentId{Id: mContent.ID}, nil
}

func (c *contentsServer) Update(ctx context.Context, content *pb.Content) (*pb.Empty, error) {
	userID, err := getUserIdFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("contentServ.Update: getUserIdFromContext: %w", err)
	}
	mContent := &model.Content{
		ID:     content.Id,
		UserID: userID,
		Type:   enum.ContentType(content.Type),
		Data:   content.Data,
		Meta:   content.Meta,
	}

	if err = c.app.Update(ctx, mContent); err != nil {
		return nil, status.Errorf(codes.Internal, "contentServ.Update: %s", err.Error())
	}
	return &pb.Empty{}, nil
}

func (c *contentsServer) Delete(ctx context.Context, cont *pb.ContentId) (*pb.Empty, error) {
	if err := c.app.Delete(ctx, cont.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "contentServ.Delete: %s", err.Error())
	}
	return &pb.Empty{}, nil
}

func (c *contentsServer) Get(ctx context.Context, cont *pb.ContentId) (*pb.Content, error) {
	res, err := c.app.Get(ctx, cont.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "contentServ.Get: %s", err.Error())
	}

	return &pb.Content{
		Id:   res.ID,
		Type: pb.TYPE(res.Type),
		Meta: res.Meta,
		Data: res.Data,
	}, nil
}

func (c *contentsServer) GetByType(q *pb.Query, s pb.Contents_GetByTypeServer) error {
	t := enum.ContentType(q.ContentType)

	contents, err := c.app.GetUserContent(s.Context(), t)
	if err != nil {
		return status.Errorf(codes.Internal, "contentServ.GetByType: %s", err.Error())
	}

	for _, content := range contents {
		err := s.Send(&pb.Content{
			Id:   content.ID,
			Type: pb.TYPE(content.Type),
			Meta: content.Meta,
			Data: content.Data,
		})
		if err != nil {
			return status.Errorf(codes.Internal, "contentServ.GetByType: s.Send:  %s", err.Error())
		}
	}

	return nil
}

//todo	SaveFile(Contents_SaveFileServer) error
//todo	GetFile(*ContentId, Contents_GetFileServer) error

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
