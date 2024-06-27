package content

import (
	"context"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
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

//TODO
