package user

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_NewAuthServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := NewMockRepository(ctrl)
	type expected struct {
		err error
	}

	tests := []struct {
		name  string
		input context.Context
		exp   expected
	}{
		{
			name:  "negative - no context",
			input: nil,
			exp: expected{
				err: errs.ErrNoCtx,
			},
		},
		{
			name:  "negative - empty context",
			input: context.TODO(),
			exp: expected{
				err: errs.ErrNoCtxLogger,
			},
		},

		{
			name:  "positive",
			input: initContext(),
			exp: expected{
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewApp(tt.input, m)
			if tt.exp.err != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.exp.err)
				assert.Empty(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
		})
	}
}

func Test_Register(t *testing.T) {
	ctx := initContext()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type expected struct {
		isError    bool
		errMessage string
		id         int32
	}

	type mockParams struct {
		id  int32
		err error
	}

	tests := []struct {
		name string
		user *model.User
		exp  expected
		mock mockParams
	}{
		{
			name: "negative - bcrypt error",
			user: &model.User{
				Login:    "login",
				Password: "этооченьоченьмногобайточеньоченьочень",
			},
			exp: expected{
				isError:    true,
				errMessage: "bcrypt: password length exceeds 72 bytes",
			},
			mock: mockParams{},
		},
		{
			name: "negative - repo error",
			user: &model.User{
				Login:    "login",
				Password: "password",
			},
			exp: expected{
				isError:    true,
				errMessage: "User.Register: repo error",
			},
			mock: mockParams{
				err: errors.New("repo error"),
			},
		},
		{
			name: "positive",
			user: &model.User{
				Login:    "login",
				Password: "password",
			},
			exp: expected{
				isError:    false,
				errMessage: "",
				id:         1,
			},
			mock: mockParams{
				id:  1,
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRepository(ctrl)
			m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(tt.mock.id, tt.mock.err).AnyTimes()
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			err = a.Register(ctx, tt.user)
			if tt.exp.isError {
				assert.ErrorContains(t, err, tt.exp.errMessage)
				assert.Empty(t, tt.user.ID)
				return
			}
			assert.Empty(t, err)
			assert.Equal(t, tt.exp.id, tt.user.ID)
			assert.NotEmpty(t, tt.user.HashPassword)
		})
	}
}

func Test_GetByLogin(t *testing.T) {
	ctx := initContext()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type expected struct {
		isError    bool
		errMessage string
	}

	type mockParams struct {
		user *model.User
		err  error
	}

	tests := []struct {
		name  string
		login string
		exp   expected
		mock  mockParams
	}{
		{
			name:  "negative",
			login: "login",
			mock: mockParams{
				user: nil,
				err:  errors.New("repo error"),
			},
			exp: expected{
				isError:    true,
				errMessage: "User.GetByLogin:",
			},
		},
		{
			name:  "positive",
			login: "login",
			mock:  mockParams{user: &model.User{ID: 1, Login: "login"}},
			exp:   expected{isError: false, errMessage: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRepository(ctrl)
			m.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(tt.mock.user, tt.mock.err).AnyTimes()
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			actUser, actErr := a.GetByLogin(ctx, tt.login)
			if tt.exp.isError {
				assert.ErrorContains(t, actErr, tt.exp.errMessage)
				return
			}
			assert.NotEmpty(t, actUser)
		})
	}
}

func Test_ValidatePassword(t *testing.T) {
	ctx := initContext()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := NewMockRepository(ctrl)

	type input struct {
		user     *model.User
		password string
	}
	type expected struct {
		isError bool
		// errMessage string
		err error
	}

	tests := []struct {
		name string
		in   input
		exp  expected
	}{
		{
			name: "negative - no user",
			in: input{
				user:     nil,
				password: "",
			},
			exp: expected{
				isError: true,
				err:     ErrValidateUser,
			},
		},
		{
			name: "negative - no hashpasword ",
			in: input{
				user:     &model.User{HashPassword: []byte{}},
				password: "",
			},
			exp: expected{
				isError: true,
				err:     ErrValidateHashPasword,
			},
		},
		{
			name: "negative - no password ",
			in: input{
				user:     &model.User{HashPassword: []byte("notempty")},
				password: "",
			},
			exp: expected{
				isError: true,
				err:     ErrValidatePasword,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			ok, err := a.ValidatePassword(tt.in.user, tt.in.password)

			if tt.exp.isError {
				assert.ErrorIs(t, err, tt.exp.err)
				assert.False(t, ok)
				return
			}
			assert.True(t, ok)

		})
	}
}

func loggerInit() (*zap.SugaredLogger, error) {
	zl, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("loggerInit: %w", err)
	}
	return zl.Sugar(), nil
}

func initContext() context.Context {
	log, err := loggerInit()
	if err != nil {
		log.Fatal(err)
	}
	return context.WithValue(context.Background(), consts.LoggerCtxKey, log)
}
