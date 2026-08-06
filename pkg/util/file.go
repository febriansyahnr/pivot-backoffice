package util

import (
	"encoding/csv"
	"mime/multipart"
	"net/http"
	"path"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type ValidateFileFormParams struct {
	FieldName string
	Extension string
	FileSize  int64
}

func ValidateFile(r *http.Request, params ValidateFileFormParams) (multipart.File, *multipart.FileHeader, error) {
	file, fileHeader, err := r.FormFile(params.FieldName)
	if err != nil {
		return nil, nil, err
	}
	if fileHeader == nil || file == nil {
		return nil, nil, http.ErrMissingFile
	}

	if !strings.EqualFold(path.Ext(fileHeader.Filename), params.Extension) {
		return nil, nil, constant.ErrInvalidFileExtension
	}

	if fileHeader.Size > params.FileSize {
		return nil, nil, constant.ErrFileTooLarge
	}

	return file, fileHeader, nil
}

type ReadCsvFileParams struct {
	IgnoreFirstRow   bool
	Delimiter        rune // ','  ';' etc
	LazyQuotes       bool
	TrimLeadingSpace bool
}

func ReadCsvFile(file multipart.File, params ReadCsvFileParams) ([][]string, error) {
	csvReader := csv.NewReader(file)
	csvReader.Comma = params.Delimiter
	csvReader.LazyQuotes = params.LazyQuotes
	csvReader.TrimLeadingSpace = params.TrimLeadingSpace

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if params.IgnoreFirstRow {
		if len(records) > 1 {
			return records[1:], nil
		} else {
			return [][]string{}, nil
		}
	}

	return records, nil
}
