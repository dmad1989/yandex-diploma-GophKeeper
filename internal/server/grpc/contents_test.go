package grpc

import (
	"testing"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/test"
	"github.com/stretchr/testify/assert"
)

func TestNewContentsServer(t *testing.T) {
	ctx, err := test.NewContextFull()
	assert.NoError(t, err)
	assert.NotEmpty(t, ctx)
	assert.NotEmpty(t, ctx.Value(consts.UserCtxKey))
}

func TestSave(t *testing.T) {

}

func TestUpdate(t *testing.T) {

}

func TestDelete(t *testing.T) {

}

func TestGet(t *testing.T) {

}

func TestGetByType(t *testing.T) {

}

func TestSaveFile(t *testing.T) {

}

func TestGetByGetFile(t *testing.T) {

}
