package xlsx

import (
	"bytes"
	"io"

	"github.com/xuri/excelize/v2"
)

type excel struct{}

type File = excelize.File
type Options = excelize.Options

type Exceler interface {
	OpenReader(r io.Reader, opts ...Options) (Filer, error)
}

type Filer interface {
	Close() error

	GetRows(sheet string, opts ...Options) ([][]string, error)
	NewSheet(sheet string) (int, error)
	SetSheetRow(sheet, cell string, slice interface{}) error
	WriteToBuffer() (*bytes.Buffer, error)
	GetCellValue(sheet, cell string, opts ...Options) (string, error)
}
