package grpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/dmad1989/gophKeeper/internal/server/grpc/mocks"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_NewAuthServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockUserApp(ctrl)
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
			res, err := NewAuthServer(tt.input, m)
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

	type mockRegisterParams struct {
		err   error
		times int
	}
	type mockGenTokenParams struct {
		token string
		err   error
		times int
	}

	type expected struct {
		errValidate error
		isError     bool
		errCode     codes.Code
	}

	tests := []struct {
		name  string
		input *gen.AuthData
		exp   expected
		mr    mockRegisterParams
		mgt   mockGenTokenParams
	}{
		{
			name:  "positive",
			input: &gen.AuthData{Username: "username", Password: "password"},
			mr:    mockRegisterParams{err: nil, times: 1},
			mgt:   mockGenTokenParams{token: "token", err: nil, times: 1},
			exp: expected{
				isError: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockUserApp(ctrl)
			m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(tt.mr.err).MaxTimes(tt.mr.times)
			m.EXPECT().GenerateToken(gomock.Any(), gomock.Any()).Return(tt.mgt.token, tt.mgt.err).MaxTimes(tt.mgt.times)
			s, err := NewAuthServer(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			res, err := s.Register(ctx, tt.input)

			if tt.exp.isError {
				assert.Error(t, err)
				if tt.exp.errValidate != nil {
					assert.ErrorIs(t, err, tt.exp.errValidate)
				}
				if e, ok := status.FromError(err); ok {
					assert.Equal(t, tt.exp.errCode, e.Code())
				}
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
			assert.NotEmpty(t, res.Token)
			assert.NotEmpty(t, res.ExpireAt)
		})
	}
}

func Test_Login(t *testing.T) {
	// ctx := context.Background()
	// ctrl := gomock.NewController(t)
	// defer ctrl.Finish()

	// tests := []struct {
	// 	name string
	// 	id   int32
	// 	exp  expected
	// 	mock mockParams
	// }{}

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		m := mocks.NewMockUserApp(ctrl)
	// 		s := NewAuthServer(ctx, m)
	// 	})
	// }
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
