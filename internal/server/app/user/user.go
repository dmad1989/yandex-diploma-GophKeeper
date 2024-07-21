package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dmad1989/gophKeeper/pkg/logging"
	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=user.go -destination=./mock_user.go -package=user

const (
	secretKey = "wbhhFFd72C4gsecretkey"
)

var (
	ErrValidateUser        = errors.New("user struct is empty")
	ErrValidatePasword     = errors.New("password  is empty")
	ErrValidateHashPasword = errors.New("hash password  is empty")
)

type Repository interface {
	CreateUser(ctx context.Context, u *model.User) (int32, error)
	GetUser(ctx context.Context, login string) (*model.User, error)
}

type UserApp struct {
	log  *zap.SugaredLogger
	repo Repository
}

type Claims struct {
	jwt.RegisteredClaims
	ID int32
}

func NewApp(ctx context.Context, r Repository) (*UserApp, error) {
	l, err := logging.LoggerFromContext(ctx, "UserApp")
	if err != nil {
		return nil, fmt.Errorf("auth.NewAuthServer: LoggerFromContext: %w", err)
	}
	return &UserApp{log: l, repo: r}, nil
}

func (a *UserApp) Register(ctx context.Context, user *model.User) error {
	var err error
	user.HashPassword, err = bcrypt.GenerateFromPassword([]byte(user.Password), 15)
	if err != nil {
		return fmt.Errorf("User.Register: crypt.GenerateFromPassword: %w", err)

	}
	user.ID, err = a.repo.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("User.Register: %w", err)
	}
	return nil
}

func (a *UserApp) GetByLogin(ctx context.Context, login string) (user *model.User, err error) {
	user, err = a.repo.GetUser(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("User.GetByLogin: %w", err)
	}
	return
}

func (a *UserApp) ValidatePassword(user *model.User, password string) (bool, error) {
	if user == nil {
		return false, ErrValidateUser
	}
	if len(user.HashPassword) == 0 {
		return false, ErrValidateHashPasword
	}
	if password == "" {
		return false, ErrValidatePasword
	}
	if err := bcrypt.CompareHashAndPassword(user.HashPassword, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("User.ValidatePassword: bcrypt.CompareHashAndPassword  %w", err)
	}
	return true, nil
}

func (a *UserApp) GenerateToken(id int32, expiredAt time.Time) (string, error) {
	if id == 0 {
		return "", errors.New("User.GenerateToken: user id is 0")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expiredAt),
			},
			ID: id,
		})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("User.GenerateToken: token.SignedString: %w", err)
	}
	return tokenString, nil
}

func (a *UserApp) ExtractIDFromToken(t string) (int32, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(t, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

	if err != nil {
		return 0, fmt.Errorf("user.ExtractIDFromToken: jwt.ParseWithClaims: %w", err)
	}
	if !token.Valid {
		return 0, fmt.Errorf("user.ExtractIDFromToken: token.Valid: %w", errs.ErrTokenInvalid)
	}
	if claims.ID == 0 {
		return 0, fmt.Errorf("user.ExtractIDFromToken: claims.ID = 0: %w", errs.ErrTokenNoUser)
	}

	return claims.ID, nil
}
