package crypto

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"go.uber.org/zap"
)

const blockLength = 128

type Configer interface {
	GetPrivateKey() *rsa.PrivateKey
}

type cryptoApp struct {
	log        *zap.SugaredLogger
	privateKey *rsa.PrivateKey
}

func New(ctx context.Context, cfg Configer) *cryptoApp {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("cryptoApp")
	return &cryptoApp{log: l, privateKey: cfg.GetPrivateKey()}
}

func (c *cryptoApp) Decrypt(data []byte) ([]byte, error) {
	if c.privateKey == nil {
		return data, nil
	}
	decryptedData := make([]byte, 0, len(data))
	var nextBlockLength int
	for i := 0; i < len(data); i += c.privateKey.PublicKey.Size() {
		nextBlockLength = i + c.privateKey.PublicKey.Size()
		if nextBlockLength > len(data) {
			nextBlockLength = len(data)
		}
		block, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.privateKey, data[i:nextBlockLength], []byte("gohkeeper"))
		if err != nil {
			return nil, fmt.Errorf("cryptoApp.Decrypt: rsa.DecryptOAEP: %w", err)
		}
		decryptedData = append(decryptedData, block...)
	}
	return decryptedData, nil
}

func (c *cryptoApp) Encrypt(data []byte) ([]byte, error) {
	if c.privateKey == nil {
		return data, nil
	}
	encryptedData := make([]byte, 0, len(data))
	var nextBlockLength int
	for i := 0; i < len(data); i += blockLength {
		nextBlockLength = i + blockLength
		if nextBlockLength > len(data) {
			nextBlockLength = len(data)
		}
		block, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &c.privateKey.PublicKey, data[i:nextBlockLength], []byte("gohkeeper"))
		if err != nil {
			return nil, fmt.Errorf("cryptoApp.Encrypt: rsa.EncryptOAEP: %w", err)
		}
		encryptedData = append(encryptedData, block...)
	}
	return encryptedData, nil
}
