package content

import (
	"context"
	"fmt"
	"io"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/client/contents"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Crypto interface {
	Decrypt(data []byte) ([]byte, error)
	Encrypt(data []byte) ([]byte, error)
}

type contentApp struct {
	log    *zap.SugaredLogger
	client pb.ContentsClient
	crypto Crypto
}

func New(ctx context.Context, conn *grpc.ClientConn, c Crypto) *contentApp {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("ContentApp")
	return &contentApp{log: l, client: pb.NewContentsClient(conn), crypto: c}
}

func (a contentApp) Save(ctx context.Context, conType enum.ContentType, data []byte, meta string) (int32, error) {
	eData, err := a.crypto.Encrypt(data)
	if err != nil {
		return 0, fmt.Errorf("ContentApp.Save: %w", err)
	}

	res, err := a.client.Save(ctx, &pb.Content{
		Type: pb.TYPE(conType),
		Data: eData,
		Meta: meta,
	})
	if err != nil {
		return 0, fmt.Errorf("ContentApp.Save: client.Save: %w", err)
	}
	return res.GetId(), nil
}
func (a contentApp) Delete(ctx context.Context, id int32) error {
	_, err := a.client.Delete(ctx, &pb.ContentId{Id: id})
	if err != nil {
		return fmt.Errorf("ContentApp.Delete: %w", err)
	}
	return nil
}
func (a contentApp) Update(ctx context.Context, id int32, contype enum.ContentType, data []byte, meta string) error {
	eData, err := a.crypto.Encrypt(data)
	if err != nil {
		return fmt.Errorf("ContentApp.Update: %w", err)
	}

	_, err = a.client.Update(ctx, &pb.Content{
		Id:   id,
		Type: pb.TYPE(contype),
		Data: eData,
		Meta: meta,
	})
	if err != nil {
		return fmt.Errorf("ContentApp.Update: client.Update: %w", err)
	}
	return nil
}
func (a contentApp) GetByType(ctx context.Context, contype enum.ContentType) ([]*model.Content, error) {
	s, err := a.client.GetByType(ctx, &pb.Query{ContentType: pb.TYPE(contype)})
	if err != nil {
		return nil, fmt.Errorf("ContentApp.GetByType: client.GetByType: %w", err)
	}
	results := make([]*model.Content, 0)
	for {
		c, err := s.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ContentApp.GetByType: loop s.Recv(): %w", err)
		}
		results = append(results, &model.Content{
			ID:   c.Id,
			Meta: c.Meta,
			Type: enum.ContentType(c.Type),
		})
	}
	return results, nil
}
func (a contentApp) Get(ctx context.Context, id int32) (*contents.Item, error) {
	content, err := a.client.Get(ctx, &pb.ContentId{Id: id})
	if err != nil {
		return nil, fmt.Errorf("ContentApp.Get: client.Get: %w", err)
	}
	decryptedData, err := a.crypto.Decrypt(content.Data)
	if err != nil {
		return nil, fmt.Errorf("ContentApp.Get: %w", err)
	}
	content.Data = decryptedData
	res, err := contents.NewItem(content)
	if err != nil {
		return nil, fmt.Errorf("ContentApp.Get: %w", err)
	}
	return res, nil
}

func (a contentApp) SaveFile(ctx context.Context, path, meta string) (int32, error) {
	//todo
	return 0, nil
}

func (a contentApp) GetFile(ctx context.Context, id int32) (string, error) {
	// todo
	return "", nil
}
