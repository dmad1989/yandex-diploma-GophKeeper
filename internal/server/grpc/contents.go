package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/dmad1989/gophKeeper/pkg/file"
	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileWorker interface {
	Save(path string, ch chan []byte) (chan error, error)
	Read(path string, errCh chan error) (chan []byte, os.FileInfo, error)
}

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
	app  ContentApp
	file FileWorker
}

func NewContentsServer(ctx context.Context, c ContentApp) pb.ContentsServer {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("ContentsServer")
	f := file.New(ctx)
	return &contentsServer{log: l, app: c, file: f}
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

func (c *contentsServer) SaveFile(s pb.Contents_SaveFileServer) error {
	ctx := s.Context()
	userID, err := getUserIdFromContext(ctx)
	if err != nil {
		return fmt.Errorf("contentServ.SaveFile: getUserIdFromContext: %w", err)
	}

	chunk, err := s.Recv()
	if err != nil {
		if err == io.EOF {
			c.log.Errorf("failed to save file content for '%d' user: empty stream", userID)
			return status.Error(codes.InvalidArgument, "empty stream")
		}
		return status.Error(codes.Internal, err.Error())
	}

	mContent := &model.Content{
		UserID: userID,
		Type:   enum.File,
		Meta:   chunk.Meta,
		Desc:   chunk.Data,
	}

	fileData := bytes.Buffer{}

	for {
		c.log.Debug("contentServ.SaveFile: start waiting stream data")
		chunk, err = s.Recv()
		if err == io.EOF {
			c.log.Debug("contentServ.SaveFile: end of stream data")
			break
		}
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("contentServ.SaveFile: s.Recv(): %w", err).Error())
		}

		_, err = fileData.Write(chunk.Data)

		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("contentServ.SaveFile: fileData.Write(chunk.Data): %w", err).Error())
		}
	}
	mContent.Data = fileData.Bytes()

	if err := c.app.Save(ctx, mContent); err != nil {
		return status.Errorf(codes.Internal, "contentServ.SaveFile: %s", err.Error())
	}

	return s.SendAndClose(&pb.ContentId{Id: mContent.ID})
}

func (c *contentsServer) GetFile(cID *pb.ContentId, s pb.Contents_GetFileServer) error {
	ctx := s.Context()
	content, err := c.app.Get(ctx, cID.Id)
	if err != nil {
		return status.Error(codes.Internal, fmt.Errorf("contentServ.GetFile:  %w", err).Error())
	}

	chunk := &pb.FileChunk{
		Meta: content.Meta,
		Data: content.Desc,
	}

	err = s.Send(chunk)
	if err != nil {
		return status.Error(codes.Internal, fmt.Errorf("contentServ.GetFile: s.Send: %w", err).Error())
	}

	buffer := make([]byte, 1024)
	chunks := bytes.SplitAfter(content.Data, buffer)

	for _, bChunk := range chunks {
		chunk = &pb.FileChunk{
			Meta: "",
			Data: bChunk,
		}
		err := s.Send(chunk)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("contentServ.GetFile: s.Send: %w", err).Error())
		}
	}
	return nil

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
