package content

import (
	"errors"
	"testing"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"github.com/dmad1989/gophKeeper/test"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ErrRepoMock = errors.New("any repo problem")
)

func TestNewApp(t *testing.T) {
	ectx := test.NewContextEmpty()
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestNewApp.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := NewMockRepository(ctrl)
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

func TestSave(t *testing.T) {
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestSave.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type expected struct {
		errValidate error
		errRepo     error
	}

	type mockParams struct {
		id  int32
		err error
	}

	tests := []struct {
		name    string
		content *model.Content
		exp     expected
		mock    mockParams
	}{
		{
			name:    "negative - no data",
			content: &model.Content{},
			exp: expected{
				errValidate: ErrEmptyData,
				errRepo:     nil,
			},
			mock: mockParams{
				id:  1,
				err: nil,
			},
		},
		{
			name: "negative - no meta",
			content: &model.Content{
				Data: []byte("data"),
			},
			exp: expected{
				errValidate: ErrEmptyMeta,
				errRepo:     nil,
			},
			mock: mockParams{
				id:  1,
				err: nil,
			},
		},

		{
			name: "negative - no type",
			content: &model.Content{
				Data: []byte("data"),
				Meta: "Meta",
			},
			exp: expected{
				errValidate: ErrEmptyType,
				errRepo:     nil,
			},
			mock: mockParams{
				id:  1,
				err: nil,
			},
		},

		{
			name: "negative - no userid",
			content: &model.Content{
				Data: []byte("data"),
				Meta: "Meta",
				Type: enum.LoginPassword,
			},
			exp: expected{
				errValidate: ErrEmptyUserID,
				errRepo:     nil,
			},
			mock: mockParams{
				id:  1,
				err: nil,
			},
		},
		{
			name: "negative - no desc for type ",
			content: &model.Content{
				UserID: 1,
				Data:   []byte("data"),
				Type:   enum.ContentType(8),
				Desc:   []byte("Desc"),
				Meta:   "Meta",
			},
			exp: expected{
				errValidate: ErrTypeToDesc,
				errRepo:     nil,
			},
			mock: mockParams{
				id:  1,
				err: nil,
			},
		},

		{
			name: "negative - repo error",
			content: &model.Content{
				UserID: 1,
				Data:   []byte("data"),
				Type:   enum.BankCard,
				Desc:   []byte("Desc"),
				Meta:   "Meta",
			},
			exp: expected{
				errValidate: nil,
				errRepo:     ErrRepoMock,
			},
			mock: mockParams{
				id:  0,
				err: ErrRepoMock,
			},
		},
		{
			name: "positive",
			content: &model.Content{
				UserID: 1,
				Data:   []byte("data"),
				Type:   enum.LoginPassword,
				Desc:   []byte("Desc"),
				Meta:   "Meta",
			},
			exp: expected{
				errValidate: nil,
				errRepo:     nil,
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
			m.EXPECT().SaveContent(gomock.Any(), gomock.Any()).Return(tt.mock.id, tt.mock.err).AnyTimes()
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			err = a.Save(ctx, tt.content)

			if tt.exp.errValidate != nil {
				assert.ErrorIs(t, err, tt.exp.errValidate)
				assert.Empty(t, tt.content.ID)
				return
			}
			if tt.exp.errRepo != nil {
				assert.ErrorIs(t, err, tt.exp.errRepo)
				assert.Empty(t, tt.content.ID)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, tt.content.ID)

		})
	}
}

func TestUpdate(t *testing.T) {
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestUpdate.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type expected struct {
		errValidate error
		errRepo     error
	}

	type mockParams struct {
		err error
	}

	tests := []struct {
		name    string
		content *model.Content
		exp     expected
		mock    mockParams
	}{
		{
			name:    "negative - no data",
			content: &model.Content{},
			exp: expected{
				errValidate: ErrEmptyData,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			},
		},
		{
			name: "negative - no meta",
			content: &model.Content{
				Data: []byte("data"),
			},
			exp: expected{
				errValidate: ErrEmptyMeta,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			},
		},

		{
			name: "negative - no type",
			content: &model.Content{
				Data: []byte("data"),
				Meta: "Meta",
			},
			exp: expected{
				errValidate: ErrEmptyType,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			},
		},

		{
			name: "negative - no userid",
			content: &model.Content{
				Data: []byte("data"),
				Meta: "Meta",
				Type: enum.LoginPassword,
			},
			exp: expected{
				errValidate: ErrEmptyUserID,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			},
		},
		{
			name: "negative - no id",
			content: &model.Content{
				UserID: 1,
				Data:   []byte("data"),
				Type:   enum.LoginPassword,
				Desc:   []byte("Desc"),
				Meta:   "Meta",
			},
			exp: expected{
				errValidate: ErrEmptyID,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			}},
		{
			name: "negative - repo error",
			content: &model.Content{
				ID:     1,
				UserID: 1,
				Data:   []byte("data"),
				Type:   enum.BankCard,
				Desc:   []byte("Desc"),
				Meta:   "Meta",
			},
			exp: expected{
				errValidate: nil,
				errRepo:     ErrRepoMock,
			},
			mock: mockParams{
				err: ErrRepoMock,
			},
		},
		{
			name: "positive",
			content: &model.Content{
				ID:     1,
				UserID: 1,
				Data:   []byte("data"),
				Type:   enum.LoginPassword,
				Desc:   []byte("Desc"),
				Meta:   "Meta",
			},
			exp: expected{
				errValidate: nil,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRepository(ctrl)
			m.EXPECT().UpdateContent(gomock.Any(), gomock.Any()).Return(tt.mock.err).AnyTimes()
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			err = a.Update(ctx, tt.content)

			if tt.exp.errValidate != nil {
				assert.ErrorIs(t, err, tt.exp.errValidate)
				return
			}
			if tt.exp.errRepo != nil {
				assert.ErrorIs(t, err, tt.exp.errRepo)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestDelete(t *testing.T) {
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestDelete.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type expected struct {
		errValidate error
		errRepo     error
	}

	type mockParams struct {
		err error
	}

	tests := []struct {
		name string
		id   int32
		exp  expected
		mock mockParams
	}{
		{
			name: "negative - no id",
			id:   0,
			exp: expected{
				errValidate: ErrEmptyID,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			}},
		{
			name: "negative - repo error",
			id:   1,
			exp: expected{
				errValidate: nil,
				errRepo:     ErrRepoMock,
			},
			mock: mockParams{
				err: ErrRepoMock,
			}},
		{
			name: "positive",
			id:   1,
			exp: expected{
				errValidate: nil,
				errRepo:     nil,
			},
			mock: mockParams{
				err: nil,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRepository(ctrl)
			m.EXPECT().DeleteContent(gomock.Any(), gomock.Any()).Return(tt.mock.err).AnyTimes()
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			err = a.Delete(ctx, tt.id)

			if tt.exp.errValidate != nil {
				assert.ErrorIs(t, err, tt.exp.errValidate)
				return
			}
			if tt.exp.errRepo != nil {
				assert.ErrorIs(t, err, tt.exp.errRepo)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetUserContent(t *testing.T) {
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestGetUserContent.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	positiveRes := make([]*model.Content, 1)
	type expected struct {
		errRepo error
	}

	type mockMethod struct {
		res   []*model.Content
		err   error
		times int
	}

	type mockParams struct {
		all  mockMethod
		byId mockMethod
	}

	tests := []struct {
		name   string
		typeID enum.ContentType
		exp    expected
		mock   mockParams
	}{
		{
			name:   "negative - by all - repo error",
			typeID: enum.Nan,
			exp: expected{
				errRepo: ErrRepoMock,
			},
			mock: mockParams{
				all: mockMethod{
					res:   nil,
					err:   ErrRepoMock,
					times: 1,
				},
				byId: mockMethod{times: 0},
			},
		},
		{
			name:   "negative - by id - repo error",
			typeID: enum.BankCard,
			exp: expected{
				errRepo: ErrRepoMock,
			},
			mock: mockParams{
				all: mockMethod{
					times: 0,
				},
				byId: mockMethod{
					res:   nil,
					err:   ErrRepoMock,
					times: 1,
				},
			},
		},

		{
			name:   "positive - by all",
			typeID: enum.Nan,
			exp: expected{
				errRepo: nil,
			},
			mock: mockParams{
				all: mockMethod{
					res:   positiveRes,
					err:   nil,
					times: 1,
				},
				byId: mockMethod{times: 0},
			},
		},
		{
			name:   "positive - by id",
			typeID: enum.File,
			exp: expected{
				errRepo: nil,
			},
			mock: mockParams{
				all: mockMethod{
					times: 0,
				},
				byId: mockMethod{
					res:   positiveRes,
					err:   nil,
					times: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRepository(ctrl)
			m.EXPECT().GetAllUserContent(gomock.Any()).Return(tt.mock.all.res, tt.mock.all.err).Times(tt.mock.all.times)
			m.EXPECT().GetUserContentByType(gomock.Any(), gomock.Any()).Return(tt.mock.byId.res, tt.mock.byId.err).Times(tt.mock.byId.times)
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			_, err = a.GetUserContent(ctx, tt.typeID)
			if tt.exp.errRepo != nil {
				assert.ErrorIs(t, err, tt.exp.errRepo)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGet(t *testing.T) {
	ctx, err := test.NewContextFull()
	require.NoError(t, err, "TestGet.NewContextFull")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type expected struct {
		errValidate error
		errRepo     error
	}

	type mockParams struct {
		content *model.Content
		err     error
	}

	tests := []struct {
		name string
		id   int32
		exp  expected
		mock mockParams
	}{
		{
			name: "negative - repo error",
			id:   1,
			exp: expected{
				errValidate: nil,
				errRepo:     ErrRepoMock,
			},
			mock: mockParams{
				content: nil,
				err:     ErrRepoMock,
			},
		},
		{
			name: "negative - no id",
			id:   0,
			exp: expected{
				errValidate: ErrEmptyID,
				errRepo:     nil,
			},
			mock: mockParams{
				content: &model.Content{},
				err:     nil,
			},
		},
		{
			name: "positive",
			id:   1,
			exp: expected{
				errValidate: nil,
				errRepo:     nil,
			},
			mock: mockParams{
				content: &model.Content{},
				err:     nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRepository(ctrl)
			m.EXPECT().GetUserContentByID(gomock.Any(), gomock.Any()).Return(tt.mock.content, tt.mock.err).AnyTimes()
			a, err := NewApp(ctx, m)
			assert.NoError(t, err, "wrong context for test! Check your realization!")
			_, err = a.Get(ctx, tt.id)
			if tt.exp.errValidate != nil {
				assert.ErrorIs(t, err, tt.exp.errValidate)
				return
			}
			if tt.exp.errRepo != nil {
				assert.ErrorIs(t, err, tt.exp.errRepo)
				return
			}
			assert.NoError(t, err)
		})
	}
}
