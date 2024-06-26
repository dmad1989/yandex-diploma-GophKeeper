package grpc

import (
	"context"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
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
	mContent := &model.Content{Type: int32(content.Type)}
	c.app.Save(ctx)
}
func (c *contentsServer) Delete(ctx context.Context, contId *pb.ContentId) (*pb.Empty, error) {}
func (c *contentsServer) Update(ctx context.Context, content *pb.Content) (*pb.Empty, error)  {}

// todo func (c *contentsServer) GetByType(q *Query, a pb.Contents_GetByTypeServer) error {}
func (c *contentsServer) Get(ctx context.Context, contId *pb.ContentId) (*pb.Content, error) {}

//todo	SaveFile(Contents_SaveFileServer) error
//todo	GetFile(*ContentId, Contents_GetFileServer) error
