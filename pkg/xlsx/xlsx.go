package xlsx

import (
	"io"

	"github.com/xuri/excelize/v2"
)

func New() Exceler {
	return &excel{}
}

func (e *excel) OpenReader(r io.Reader, opts ...Options) (Filer, error) {
	return excelize.OpenReader(r, opts...)
}
