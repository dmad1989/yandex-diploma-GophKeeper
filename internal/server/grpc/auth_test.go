package grpc

import (
	"errors"
	"testing"

	"github.com/dmad1989/gophKeeper/internal/server/grpc/mocks"
	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"github.com/dmad1989/gophKeeper/test"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errUserAppMock = errors.New("user app error")

	authDataFull       = gen.AuthData{Username: "Username", Password: "password"}
	authDataNoLogin    = gen.AuthData{Username: "", Password: "password"}
	authDataNoPassword = gen.AuthData{Username: "Username", Password: ""}
)

type mockGenTokenParams struct {
	token string
	err   error
	times int
}

func TestNewAuthServer(t *testing.T) {
	ectx := test.NewContextEmpty()
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestNewAuthServer.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockUserApp(ctrl)
	type expected struct {
		err error
	}

	tests := []struct {
		name  string
		input test.Context
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
			input: ectx,
			exp: expected{
				err: errs.ErrNoCtxLogger,
			},
		},

		{
			name:  "positive",
			input: ctx,
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

func TestRegister(t *testing.T) {
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestRegister.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type mockRegisterParams struct {
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
			name:  "negative - userApp.generateToken - error",
			input: &authDataFull,
			mr:    mockRegisterParams{err: nil, times: 1},
			mgt:   mockGenTokenParams{token: "", err: errUserAppMock, times: 1},
			exp: expected{
				isError:     true,
				errValidate: errUserAppMock,
				errCode:     codes.Internal,
			},
		},

		{
			name:  "negative -userApp.Register - any error",
			input: &authDataFull,
			mr:    mockRegisterParams{err: errUserAppMock, times: 1},
			mgt:   mockGenTokenParams{token: "token", err: nil, times: 0},
			exp: expected{
				isError:     true,
				errValidate: errUserAppMock,
				errCode:     codes.Internal,
			},
		},

		{
			name:  "negative - userApp.Register - already exists error",
			input: &authDataFull,
			mr:    mockRegisterParams{err: errs.ErrUserAlreadyExist, times: 1},
			mgt:   mockGenTokenParams{token: "token", err: nil, times: 0},
			exp: expected{
				isError:     true,
				errValidate: errs.ErrUserAlreadyExist,
				errCode:     codes.AlreadyExists,
			},
		},
		{
			name:  "negative - no password",
			input: &authDataNoPassword,
			mr:    mockRegisterParams{err: nil, times: 0},
			mgt:   mockGenTokenParams{token: "token", err: nil, times: 0},
			exp: expected{
				isError:     true,
				errValidate: ErrPasswordEmpty,
				errCode:     codes.InvalidArgument,
			},
		},
		{
			name:  "negative - no username",
			input: &authDataNoLogin,
			mr:    mockRegisterParams{err: nil, times: 0},
			mgt:   mockGenTokenParams{token: "token", err: nil, times: 0},
			exp: expected{
				isError:     true,
				errValidate: ErrUsernameEmpty,
				errCode:     codes.InvalidArgument,
			},
		},

		{
			name:  "positive",
			input: &authDataFull,
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
			require.NoError(t, err, "wrong context for test! Check your realization!")
			res, err := s.Register(ctx, tt.input)

			if tt.exp.isError {
				if s, ok := status.FromError(err); ok {
					assert.Equal(t, tt.exp.errCode, s.Code())
					if tt.exp.errValidate != nil {
						assert.ErrorContains(t, s.Err(), tt.exp.errValidate.Error())
					}
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

func TestLogin(t *testing.T) {
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestLogin.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type mockGetByLoginParams struct {
		err   error
		user  *model.User
		times int
	}

	type mockValidatePasswordParams struct {
		err   error
		ok    bool
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
		ml    mockGetByLoginParams
		mv    mockValidatePasswordParams
		mgt   mockGenTokenParams
	}{
		{
			name:  "negative - generateToken - error",
			input: &authDataFull,
			ml: mockGetByLoginParams{
				user:  &model.User{ID: 1},
				err:   nil,
				times: 1,
			},
			mv: mockValidatePasswordParams{
				ok:    true,
				err:   nil,
				times: 1,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   errUserAppMock,
				times: 1,
			},
			exp: expected{
				isError:     true,
				errValidate: nil,
				errCode:     codes.Internal,
			},
		},
		{
			name:  "negative - ValidatePassword - password not valid",
			input: &authDataFull,
			ml: mockGetByLoginParams{
				user:  &model.User{ID: 1},
				err:   nil,
				times: 1,
			},
			mv: mockValidatePasswordParams{
				ok:    false,
				err:   nil,
				times: 1,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   nil,
				times: 0,
			},
			exp: expected{
				isError:     true,
				errValidate: nil,
				errCode:     codes.InvalidArgument,
			},
		},
		{
			name:  "negative - ValidatePassword - other error",
			input: &authDataFull,
			ml: mockGetByLoginParams{
				user:  &model.User{ID: 1},
				err:   nil,
				times: 1,
			},
			mv: mockValidatePasswordParams{
				ok:    false,
				err:   errUserAppMock,
				times: 1,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   nil,
				times: 0,
			},
			exp: expected{
				isError:     true,
				errValidate: nil,
				errCode:     codes.Internal,
			},
		},
		{
			name:  "negative - GetByLogin - other error",
			input: &authDataFull,
			ml: mockGetByLoginParams{
				user:  nil,
				err:   errUserAppMock,
				times: 1,
			},
			mv: mockValidatePasswordParams{
				ok:    false,
				times: 0,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   nil,
				times: 0,
			},
			exp: expected{
				isError:     true,
				errValidate: nil,
				errCode:     codes.Internal,
			},
		},
		{
			name:  "negative - user not found",
			input: &authDataFull,
			ml: mockGetByLoginParams{
				user:  nil,
				err:   errs.ErrUserNotFound,
				times: 1,
			},
			mv: mockValidatePasswordParams{
				ok:    true,
				err:   nil,
				times: 0,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   nil,
				times: 0,
			},
			exp: expected{
				isError:     true,
				errValidate: errs.ErrUserNotFound,
				errCode:     codes.NotFound,
			},
		},
		{
			name:  "negative - no password",
			input: &authDataNoPassword,
			ml: mockGetByLoginParams{
				user:  nil,
				err:   nil,
				times: 0,
			},
			mv: mockValidatePasswordParams{
				ok:    true,
				err:   nil,
				times: 0,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   nil,
				times: 0,
			},
			exp: expected{
				isError:     true,
				errValidate: ErrPasswordEmpty,
				errCode:     codes.InvalidArgument,
			},
		},
		{
			name:  "negative - no username",
			input: &authDataNoLogin,
			ml: mockGetByLoginParams{
				user:  nil,
				err:   nil,
				times: 0,
			},
			mv: mockValidatePasswordParams{
				ok:    true,
				err:   nil,
				times: 0,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   nil,
				times: 0,
			},
			exp: expected{
				isError:     true,
				errValidate: ErrUsernameEmpty,
				errCode:     codes.InvalidArgument,
			},
		},

		{
			name: "positive",
			input: &gen.AuthData{
				Username: "username",
				Password: "password",
			},
			ml: mockGetByLoginParams{
				user:  &model.User{ID: 1},
				err:   nil,
				times: 1,
			},
			mv: mockValidatePasswordParams{
				ok:    true,
				err:   nil,
				times: 1,
			},
			mgt: mockGenTokenParams{
				token: "token",
				err:   nil,
				times: 1,
			},
			exp: expected{
				isError: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockUserApp(ctrl)

			m.EXPECT().GenerateToken(gomock.Any(), gomock.Any()).Return(tt.mgt.token, tt.mgt.err).MaxTimes(tt.mgt.times)
			m.EXPECT().GetByLogin(gomock.Any(), gomock.Any()).Return(tt.ml.user, tt.ml.err).MaxTimes(tt.ml.times)
			m.EXPECT().ValidatePassword(gomock.Any(), gomock.Any()).Return(tt.mv.ok, tt.mv.err).MaxTimes(tt.mv.times)

			s, err := NewAuthServer(ctx, m)
			require.NoError(t, err, "wrong context for test! Check your realization!")
			res, err := s.Login(ctx, tt.input)

			if tt.exp.isError {
				if s, ok := status.FromError(err); ok {
					assert.Equal(t, tt.exp.errCode, s.Code())
					if tt.exp.errValidate != nil {
						assert.ErrorContains(t, s.Err(), tt.exp.errValidate.Error())
					}
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
