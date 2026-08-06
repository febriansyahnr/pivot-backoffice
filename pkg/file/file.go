package file

import (
	"errors"
	"io"
	"mime/multipart"
	"os"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/tealeg/xlsx"
)

type file struct {
	fl        *os.File
	readExcel func(*os.File) (*xlsx.Sheet, error)
}

type FileFunc func(*file)

func WithReadExcel(read func(*os.File) (*xlsx.Sheet, error)) FileFunc {
	return func(f *file) {
		f.readExcel = read
	}
}

// Name returns the name of the file as presented to Open.
func (f *file) Name() string {
	return f.fl.Name()
}

func (f *file) Close() error {
	if err := f.fl.Close(); err != nil {
		return err
	}
	return os.Remove(f.fl.Name())
}

func (f *file) ReadExcel() (*xlsx.Sheet, error) {
	if f.readExcel == nil {
		return nil, errors.New("excel reader has not been defined")
	}
	return f.readExcel(f.fl)
}

func CopyTempMultipartFile(dir, pattern string, src multipart.File, opts ...FileFunc) (f *file, err error) {

	if err = util.EnsureTmpDir(dir); err != nil {
		return
	}

	f = &file{}

	if f.fl, err = os.CreateTemp(dir, pattern); err != nil {
		return nil, err
	}

	if _, err = io.Copy(f.fl, src); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(f)
	}
	return
}
