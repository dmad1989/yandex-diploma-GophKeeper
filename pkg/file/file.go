package file

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	"go.uber.org/zap"
)

const (
	bufferSize  = 655360
	maxFileSize = 655360
)

type File struct {
	log *zap.SugaredLogger
}

func New(ctx context.Context) *File {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("File")
	return &File{log: l}
}

func (f *File) Read(path string, errCh chan error) (chan []byte, os.FileInfo, error) {
	buf := make(chan []byte)
	file, err := os.Open(path)
	if err != nil {
		return buf, nil, fmt.Errorf("File.Read: os.Open(path): %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("File.Read: file.Stat(): %w", err)
	}
	if stat.Size() > maxFileSize {
		return nil, nil, errs.ErrFileMaxSize
	}
	go func() {
		defer file.Close()
		reader := bufio.NewReader(file)
		buffer := make([]byte, bufferSize)
		n := 0
		for {
			n, err = reader.Read(buffer)
			if err == io.EOF || n == 0 {
				close(buf)
				return
			}
			if err != nil {
				close(buf)
				return
			}

			select {
			case buf <- buffer[:n]:
			case <-errCh:
				close(buf)
				return
			case <-time.After(1 * time.Minute):
				f.log.Errorf("failed to read file: channel send timeout")
				close(buf)
				return
			}
		}
	}()

	return buf, stat, nil
}

func (f *File) Save(path string, ch chan []byte) (chan error, error) {
	errCh := make(chan error)
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("File.Save: os.Create(path): %w", err)
	}
	go func() {
		writer := bufio.NewWriter(file)
		defer file.Close()
		defer writer.Flush()
		for {
			if bytes, ok := <-ch; ok {
				_, err = writer.Write(bytes)
				if err != nil {
					f.log.Errorf("failed to save file: %v", err)
					errCh <- fmt.Errorf("File.Save:  writer.Write(bytes): %w", err)
					return
				}
				continue
			}
			break
		}
	}()
	return errCh, nil
}
