package mocks

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
)

func CreateMultipartFormFile(key, filename, content string) (*bytes.Buffer, string, error) {
	buff := new(bytes.Buffer)

	mw := multipart.NewWriter(buff)

	dataPart, err := mw.CreateFormFile(key, filename)
	if err != nil {
		return nil, "", err
	}

	if _, err = io.WriteString(dataPart, content); err != nil {
		return nil, "", err
	}

	if err := mw.Close(); err != nil {
		return nil, "", err
	}
	return buff, mw.FormDataContentType(), nil
}

func ResponseFormattingTest(code int, msg string, respType ...string) string {
	if len(respType) == 0 {
		respType = []string{"API_ERROR"}
	}
	return fmt.Sprintf(
		`{"code":"%d","message":"%s","error":{"type":"%s","message":"%s","recommendation":""},"data":null}`, code, msg, respType[0], msg,
	)
}
