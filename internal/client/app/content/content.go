package content

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dmad1989/gophKeeper/pkg/file"
	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/client/contents"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type CryptoWorker interface {
	Decrypt(data []byte) ([]byte, error)
	Encrypt(data []byte) ([]byte, error)
}

type FileWorker interface {
	Save(path string, ch chan []byte) (chan error, error)
	Read(path string, errCh chan error) (chan []byte, os.FileInfo, error)
}

type contentApp struct {
	log    *zap.SugaredLogger
	client pb.ContentsClient
	crypto CryptoWorker
	file   FileWorker
}

func New(ctx context.Context, conn *grpc.ClientConn, c CryptoWorker) *contentApp {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("ContentApp")
	f := file.New(ctx)
	return &contentApp{log: l, client: pb.NewContentsClient(conn), crypto: c, file: f}
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
	s, err := a.client.SaveFile(ctx)
	if err != nil {
		return 0, fmt.Errorf("ContentApp.SaveFile: %w", err)
	}
	errCh := make(chan error)

	chunks, stat, err := a.file.Read(path, errCh)

	if err != nil {
		return 0, fmt.Errorf("ContentApp.SaveFile: %w", err)
	}

	fileDesc := contents.File{
		Name:      stat.Name(),
		Extension: filepath.Ext(path),
		Size:      stat.Size()}

	fileDescJson, err := json.Marshal(fileDesc)

	if err != nil {
		return 0, fmt.Errorf("ContentApp.SaveFile: json.Marshal(file): %w", err)
	}

	err = s.Send(&pb.FileChunk{
		Meta: meta,
		Data: fileDescJson,
	})
	if err != nil {
		return 0, fmt.Errorf("ContentApp.SaveFile: s.Send Desc: %w", err)
	}

	for {
		chunk, ok := <-chunks
		if !ok {
			break
		}

		data, err := a.crypto.Encrypt(chunk)
		if err != nil {
			return 0, fmt.Errorf("ContentApp.SaveFile: %w", err)
		}

		err = s.Send(&pb.FileChunk{Data: data})
		if err != nil {
			errCh <- err
			return 0, fmt.Errorf("ContentApp.SaveFile: s.Send Data: %w", err)
		}
	}

	contId, err := s.CloseAndRecv()
	if err != nil {
		return 0, fmt.Errorf("ContentApp.SaveFile: s.CloseAndRecv(): %w", err)
	}

	return contId.Id, nil
}

func (a contentApp) GetFile(ctx context.Context, id int32) (string, error) {
	s, err := a.client.GetFile(ctx, &pb.ContentId{Id: id})
	if err != nil {
		return "", fmt.Errorf("ContentApp.GetFile: %w", err)
	}

	chunk, err := s.Recv()
	if err != nil {
		return "", fmt.Errorf("ContentApp.GetFile: s.Recv(): %w", err)
	}

	var fileDesc contents.File
	err = json.Unmarshal(chunk.Data, &fileDesc)
	if err != nil {
		return "", fmt.Errorf("ContentApp.GetFile: json.Unmarshal: %w", err)
	}

	path := fmt.Sprintf("./%s", fileDesc.Name)
	chunks := make(chan []byte)
	errCh, err := a.file.Save(path, chunks)
	if err != nil {
		return "", fmt.Errorf("ContentApp.GetFile: %w", err)
	}

Loop:
	for {
		chunk, err := s.Recv()
		if err == io.EOF {
			close(chunks)
			break Loop
		}
		if err != nil {
			close(chunks)
			return "", fmt.Errorf("ContentApp.SaveFile: s.Recv(): %w", err)
		}

		data, err := a.crypto.Decrypt(chunk.Data)
		if err != nil {
			return "", fmt.Errorf("ContentApp.SaveFile: %w", err)
		}

		select {
		case chunks <- data:
		case <-errCh:
			close(chunks)
			break Loop
		}
	}

	return path, nil
}
