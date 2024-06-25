package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrUsernameEmpty = errors.New("username is empty")
	ErrPasswordEmpty = errors.New("password is empty")
)

type UserApp interface {
	Register(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, login string) (user *model.User, err error)
	ValidatePassword(user *model.User, password string) (bool, error)
	GenerateToken(id int32, expiredAt time.Time) (string, error)
	ExtractIDFromToken(t string) (int32, error)
}

type authServ struct {
	log     *zap.SugaredLogger
	userApp UserApp
	pb.UnimplementedAuthServer
}

func NewAuthServer(ctx context.Context, u UserApp) pb.AuthServer {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("UserApp")
	return &authServ{log: l, userApp: u}
}

func (a *authServ) Register(ctx context.Context, ad *pb.AuthData) (*pb.TokenData, error) {
	err := validate(ad)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "auth.Register: %s", err.Error())
	}
	user := model.User{Login: ad.Username, Password: ad.Password}

	err = a.userApp.Register(ctx, &user)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExist) {
			return nil, status.Errorf(codes.AlreadyExists, "auth.Register: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "auth.Register: %s", err.Error())
	}
	t, err := a.generateToken(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "auth.Register: %s", err.Error())
	}
	return t, nil
}
func (a *authServ) Login(ctx context.Context, ad *pb.AuthData) (*pb.TokenData, error) {
	err := validate(ad)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "auth.Login: %s", err.Error())
	}

	user, err := a.userApp.GetByLogin(ctx, ad.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "auth.Login: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "auth.Login: %s", err.Error())
	}

	ok, err := a.userApp.ValidatePassword(user, ad.Password)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "auth.Login: %s", err.Error())
	}

	if !ok {
		return nil, status.Error(codes.InvalidArgument, "auth.Login: password is incorrect")
	}
	t, err := a.generateToken(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "auth.Login: %s", err.Error())
	}
	return t, nil
}

func (a *authServ) generateToken(id int32) (*pb.TokenData, error) {
	expireAt := time.Now().UTC().Add(time.Hour * 3)
	t, err := a.userApp.GenerateToken(id, expireAt)
	if err != nil {
		return nil, fmt.Errorf("auth.generateToken: %w", err)
	}
	return &pb.TokenData{Token: t, ExpireAt: timestamppb.New(expireAt)}, nil
}

func validate(a *pb.AuthData) error {
	if a.Username == "" {
		return ErrUsernameEmpty
	}
	if a.Password == "" {
		return ErrPasswordEmpty
	}
	return nil
}
