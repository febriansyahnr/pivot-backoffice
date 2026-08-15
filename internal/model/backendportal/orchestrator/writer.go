package orchestrator_model

import (
	"io"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type FileWriter struct {
	io.Writer
}

func (w *FileWriter) WriteHeader(filename string) {
	wr, ok := w.Writer.(http.ResponseWriter)
	if !ok {
		return
	}
	wr.Header().Set(constant.HeaderContentType, "application/octet-stream")
	wr.Header().Set("Content-Disposition", "attachment; filename="+filename)
	wr.Header().Set("Content-Transfer-Encoding", "binary")
	wr.Header().Set("Expires", "0")
}
